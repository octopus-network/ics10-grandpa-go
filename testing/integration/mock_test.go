package integration_test

import (
	"context"
	"encoding/hex"
	"log"
	"sort"
	"testing"

	"github.com/ComposableFi/go-merkle-trees/hasher"
	"github.com/ComposableFi/go-merkle-trees/merkle"
	"github.com/ComposableFi/go-merkle-trees/mmr"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"

	// "github.com/centrifuge/go-substrate-rpc-client/v4/xxhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	grandpa "github.com/octopus-network/grandpa-go/10-grandpa"
	"github.com/octopus-network/grandpa-go/util"
	"github.com/stretchr/testify/require"
)

var endpoint = "wss://rococo-rpc.polkadot.io"

// var polkadot_endpoint = "wss://rpc.polkadot.io"
// var local_endpoint = "ws://127.0.0.1:9944"

const PARA_ID uint32 = 2087

func TestUpdateClient(t *testing.T) {

	relayApi, err := gsrpc.NewSubstrateAPI(endpoint)
	require.NoError(t, err)

	t.Log("==== connected! ==== ")

	// _parachainApi, err := client.NewSubstrateAPI("wss://127.0.0.1:9988")
	// if err != nil {
	// 	panic(err)
	// }

	// channel to receive new SignedCommitments
	ch := make(chan interface{})

	sub, err := relayApi.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch,
	)
	require.NoError(t, err)

	t.Log("====== subcribed! ======")
	var clientState *grandpa.ClientState
	defer sub.Unsubscribe()

	for count := 0; count < 100; count++ {
		select {
		case msg, ok := <-ch:
			require.True(t, ok, "error reading channel")

			// compactCommitment := clientTypes.CompactSignedCommitment{}
			s := &grandpa.VersionedFinalityProof{}
			// attempt to decode the SignedCommitments
			// err = types.DecodeFromHexString(msg.(string), &compactCommitment)
			err := codec.DecodeFromHex(msg.(string), s)
			require.NoError(t, err)

			// signedCommitment := compactCommitment.Unpack()
			// latest finalized block number
			// blockNumber := uint32(signedCommitment.Commitment.BlockNumber)
			log.Printf("decoded msg: %#v\n", s)
			blockNumber := s.SignedCommitment.Commitment.BlockNumber
			log.Printf("blockNumber: %d\n", blockNumber)

			// initialize our client state
			if clientState != nil && clientState.LatestBeefyHeight >= blockNumber {
				t.Logf("Skipping stale Commitment for block: %d", s.SignedCommitment.Commitment.BlockNumber)
				continue
			}

			// convert to the blockHash
			blockHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(blockNumber))
			require.NoError(t, err)
			log.Printf("blockHash: %#v\n", codec.HexEncodeToString(blockHash[:]))

			authorities, err := util.GetBeefyAuthorities(blockHash, relayApi, "Authorities")
			require.NoError(t, err)

			nextAuthorities, err := util.GetBeefyAuthorities(blockHash, relayApi, "NextAuthorities")
			require.NoError(t, err)

			var authorityLeaves [][]byte
			for _, v := range authorities {
				authorityLeaves = append(authorityLeaves, crypto.Keccak256(v))
			}

			authorityTree, err := merkle.NewTree(hasher.Keccak256Hasher{}).FromLeaves(authorityLeaves)
			require.NoError(t, err)

			var nextAuthorityLeaves [][]byte
			for _, v := range nextAuthorities {
				nextAuthorityLeaves = append(nextAuthorityLeaves, crypto.Keccak256(v))
			}

			nextAuthorityTree, err := merkle.NewTree(hasher.Keccak256Hasher{}).FromLeaves(nextAuthorityLeaves)
			require.NoError(t, err)

			if clientState == nil {
				var authorityTreeRoot = grandpa.Bytes32(authorityTree.Root())
				var nextAuthorityTreeRoot = grandpa.Bytes32(nextAuthorityTree.Root())

				clientState = &grandpa.ClientState{
					MmrRootHash:          s.SignedCommitment.Commitment.Payload[0].Data,
					LatestBeefyHeight:    blockNumber,
					BeefyActivationBlock: 0,
					Authority: &grandpa.BeefyAuthoritySet{
						Id:            uint64(s.SignedCommitment.Commitment.ValidatorSetID),
						Len:           uint32(len(authorities)),
						AuthorityRoot: &authorityTreeRoot,
					},
					NextAuthoritySet: &grandpa.BeefyAuthoritySet{
						Id:            uint64(s.SignedCommitment.Commitment.ValidatorSetID) + 1,
						Len:           uint32(len(nextAuthorities)),
						AuthorityRoot: &nextAuthorityTreeRoot,
					},
				}
				log.Printf("Initializing client state: $%#v", clientState)
				continue
			}

			log.Printf("Recieved Signed Commitment count: #%d", count)

			// first get all paraIds
			// fetch all registered parachainIds, this method doesn't account for
			// if the parachains whose header was included in the batch of finalized blocks have now
			// lost their parachain slot at this height
			paraIds, err := util.FetchParaIDs(relayApi, blockHash)
			require.NoError(t, err)
			log.Printf("paraIds: $%#v", paraIds)

			// var paraHeaderKeys []types.StorageKey

			// // create full storage key for our own paraId
			// keyPrefix := util.CreateStorageKeyPrefix("Paras", "Heads")
			// log.Printf("keyPrefix: $%#v", codec.HexEncodeToString(keyPrefix[:]))
			// // so we can query all blocks from lastfinalized to latestBeefyHeight
			// log.Printf("PARA_ID: $%#v", PARA_ID)
			// encodedParaID, err := codec.Encode(PARA_ID)
			// log.Printf("encodedParaID: $%#v", codec.HexEncodeToString(encodedParaID[:]))
			// require.NoError(t, err)

			// twoXHash := xxhash.New64(encodedParaID).Sum(nil)
			// log.Printf("encodedParaID twoXHash: $%#v", codec.HexEncodeToString(twoXHash[:]))
			// // full key path in the storage source: https://www.shawntabrizi.com/assets/presentations/substrate-storage-deep-dive.pdf
			// // xx128("Paras") + xx128("Heads") + xx64(Encode(paraId)) + Encode(paraId)
			// fullKey := append(append(keyPrefix, twoXHash[:]...), encodedParaID...)
			// log.Printf("fullKey: $%#v", codec.HexEncodeToString(fullKey[:]))
			// paraHeaderKeys = append(paraHeaderKeys, fullKey)

			// previousFinalizedHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(clientState.LatestBeefyHeight + 1))
			// require.NoError(t, err)
			// log.Printf("previousFinalizedHash: %#v\n", codec.HexEncodeToString(previousFinalizedHash[:]))

			// changeSet, err := relayApi.RPC.State.QueryStorage(paraHeaderKeys, previousFinalizedHash, blockHash)
			// require.NoError(t, err)

			changeSet, err := util.QueryParaChainStorage(relayApi, PARA_ID, clientState.LatestBeefyHeight+1, blockNumber)
			require.NoError(t, err)

			// double map that holds block numbers, for which parachain header
			// was included in the mmr leaf, seeing as our parachain headers might not make it into
			// every relay chain block.
			// Map<BlockNumber, Map<ParaId, Header>>
			var finalizedBlocks = make(map[uint32]map[uint32][]byte)

			// request for batch mmr proof of those leaves
			var leafIndices []uint64

			for _, changes := range changeSet {
				header, err := relayApi.RPC.Chain.GetHeader(changes.Block)
				require.NoError(t, err)

				var heads = make(map[uint32][]byte)

				for _, paraId := range paraIds {
					header, err := util.FetchParachainHeader(relayApi, paraId, changes.Block)
					require.NoError(t, err)
					heads[paraId] = header
				}

				finalizedBlocks[uint32(header.Number)] = heads

				leafIndices = append(leafIndices, uint64(clientState.GetLeafIndexForBlockNumber(uint32(header.Number))))
			}

			// fetch mmr proofs for leaves containing our target paraId
			mmrBatchProof, err := util.GenerateBatchProof(relayApi, &blockHash, leafIndices)
			require.NoError(t, err)
			log.Printf("mmrBatchProof: %#v\n", mmrBatchProof)

			var parachainHeaders []*grandpa.ParachainHeader

			for i := 0; i < len(mmrBatchProof.Leaves); i++ {
				type LeafWithIndex struct {
					Leaf  types.MMRLeaf
					Index uint64
				}

				v := LeafWithIndex{Leaf: mmrBatchProof.Leaves[i], Index: uint64(mmrBatchProof.Proof.LeafIndex[i])}
				var leafBlockNumber = clientState.GetBlockNumberForLeaf(uint32(v.Index))
				paraHeaders := finalizedBlocks[leafBlockNumber]

				var paraHeadsLeaves [][]byte
				// index of our parachain header in the
				// parachain heads merkle root
				var index uint32

				count := 0

				// sort by paraId
				var sortedParaIds []uint32
				for paraId := range paraHeaders {
					sortedParaIds = append(sortedParaIds, paraId)
				}
				sort.SliceStable(sortedParaIds, func(i, j int) bool {
					return sortedParaIds[i] < sortedParaIds[j]
				})

				type ParaIdAndHeader struct {
					ParaId uint32
					Header []byte
				}

				for _, paraId := range sortedParaIds {
					bytes, err := codec.Encode(ParaIdAndHeader{ParaId: paraId, Header: paraHeaders[paraId]})
					require.NoError(t, err)
					leafHash := crypto.Keccak256(bytes)
					paraHeadsLeaves = append(paraHeadsLeaves, leafHash)
					if paraId == PARA_ID {
						// note index of paraId
						index = uint32(count)
					}
					count++
				}

				tree, err := merkle.NewTree(hasher.Keccak256Hasher{}).FromLeaves(paraHeadsLeaves)
				require.NoError(t, err)

				paraHeadsProof := tree.Proof([]uint64{uint64(index)})
				authorityRoot := grandpa.Bytes32(v.Leaf.BeefyNextAuthoritySet.Root[:])
				parentHash := grandpa.Bytes32(v.Leaf.ParentNumberAndHash.Hash[:])

				header := grandpa.ParachainHeader{
					ParachainHeader: paraHeaders[PARA_ID],
					MmrLeafPartial: &grandpa.BeefyMmrLeafPartial{
						Version:      grandpa.U8(v.Leaf.Version),
						ParentNumber: uint32(v.Leaf.ParentNumberAndHash.ParentNumber),
						ParentHash:   &parentHash,
						BeefyNextAuthoritySet: grandpa.BeefyAuthoritySet{
							Id:            uint64(v.Leaf.BeefyNextAuthoritySet.ID),
							Len:           uint32(v.Leaf.BeefyNextAuthoritySet.Len),
							AuthorityRoot: &authorityRoot,
						},
					},
					ParachainHeadsProof: paraHeadsProof.ProofHashes(),
					ParaId:              PARA_ID,
					HeadsLeafIndex:      index,
					HeadsTotalCount:     uint32(len(paraHeadsLeaves)),
				}

				parachainHeaders = append(parachainHeaders, &header)
			}

			mmrProof, err := relayApi.RPC.MMR.GenerateProof(
				uint64(clientState.GetLeafIndexForBlockNumber(blockNumber)),
				blockHash,
			)
			require.NoError(t, err)

			log.Printf("mmrProof: %#v\n", mmrProof)

			latestLeaf := mmrProof.Leaf

			BeefyNextAuthoritySetRoot := grandpa.Bytes32(latestLeaf.BeefyNextAuthoritySet.Root[:])
			parentHash := grandpa.Bytes32(latestLeaf.ParentNumberAndHash.Hash[:])

			var latestLeafMmrProof = make([][]byte, len(mmrProof.Proof.Items))
			for i := 0; i < len(mmrProof.Proof.Items); i++ {
				latestLeafMmrProof[i] = mmrProof.Proof.Items[i][:]
			}
			var mmrBatchProofItems = make([][]byte, len(mmrBatchProof.Proof.Items))
			for i := 0; i < len(mmrBatchProof.Proof.Items); i++ {
				mmrBatchProofItems[i] = mmrBatchProof.Proof.Items[i][:]
			}
			var signatures []*grandpa.CommitmentSignature
			var authorityIndices []uint64
			// luckily for us, this is already sorted and maps to the right authority index in the authority root.
			for i, v := range s.SignedCommitment.Signatures {
				if v.IsSome() {
					_, sig := v.Unwrap()
					signatures = append(signatures, &grandpa.CommitmentSignature{
						Signature:      sig[:],
						AuthorityIndex: uint32(i),
					})
					authorityIndices = append(authorityIndices, uint64(i))
				}
			}

			CommitmentPayload := s.SignedCommitment.Commitment.Payload[0]
			var payloadId grandpa.SizedByte2 = CommitmentPayload.ID
			ParachainHeads := grandpa.Bytes32(latestLeaf.ParachainHeads[:])
			leafIndex := clientState.GetLeafIndexForBlockNumber(blockNumber)

			mmrUpdateProof := grandpa.MmrUpdateProof{
				MmrLeaf: &grandpa.BeefyMmrLeaf{
					Version:        grandpa.U8(latestLeaf.Version),
					ParentNumber:   uint32(latestLeaf.ParentNumberAndHash.ParentNumber),
					ParentHash:     &parentHash,
					ParachainHeads: &ParachainHeads,
					BeefyNextAuthoritySet: grandpa.BeefyAuthoritySet{
						Id:            uint64(latestLeaf.BeefyNextAuthoritySet.ID),
						Len:           uint32(latestLeaf.BeefyNextAuthoritySet.Len),
						AuthorityRoot: &BeefyNextAuthoritySetRoot,
					},
				},
				MmrLeafIndex: uint64(leafIndex),
				MmrProof:     latestLeafMmrProof,
				SignedCommitment: &grandpa.SignedCommitment{
					Commitment: &grandpa.Commitment{
						Payload:        []*grandpa.PayloadItem{{PayloadId: &payloadId, PayloadData: CommitmentPayload.Data}},
						BlockNumer:     uint32(s.SignedCommitment.Commitment.BlockNumber),
						ValidatorSetId: uint64(s.SignedCommitment.Commitment.ValidatorSetID),
					},
					Signatures: signatures,
				},
				AuthoritiesProof: authorityTree.Proof(authorityIndices).ProofHashes(),
			}
			log.Printf("mmrUpdateProof: %#v\n", mmrUpdateProof)

			header := grandpa.Header{
				ParachainHeaders: parachainHeaders,
				MmrProofs:        mmrBatchProofItems,
				MmrSize:          mmr.LeafIndexToMMRSize(uint64(leafIndex)),
				MmrUpdateProof:   &mmrUpdateProof,
			}
			log.Printf("header: %#v\n", header)

			err = clientState.VerifyClientMessage(sdk.Context{}, nil, nil, &header)
			require.NoError(t, err)

			t.Logf("clientState.LatestBeefyHeight: %d clientState.MmrRootHash: %s", clientState.LatestBeefyHeight, hex.EncodeToString(clientState.MmrRootHash))

			if clientState.LatestBeefyHeight != uint32(s.SignedCommitment.Commitment.BlockNumber) {
				require.Equal(t, clientState.MmrRootHash, s.SignedCommitment.Commitment.Payload, "failed to update client state. LatestBeefyHeight: %d, Commitment.BlockNumber %d", clientState.LatestBeefyHeight, uint32(s.SignedCommitment.Commitment.BlockNumber))
			}
			t.Log("====== successfully processed justification! ======")

			// if UPDATE_STATE_MODE == "true" {
			// 	paramSpace := types2.NewSubspace(nil, nil, nil, nil, "test")
			// 	//paramSpace = paramSpace.WithKeyTable(clientypes.ParamKeyTable())

			// 	k := keeper.NewKeeper(nil, nil, paramSpace, nil, nil)
			// 	ctx := sdk.Context{}
			// 	store := k.ClientStore(ctx, "1234")

			// 	clientState.UpdateState(sdk.Context{}, nil, store, &header)
			// }

			// TODO: assert that the consensus states were actually persisted
			// TODO: tests against invalid proofs and consensus states
		}
	}
}
