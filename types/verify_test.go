package grandpa_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ComposableFi/go-merkle-trees/hasher"
	"github.com/ComposableFi/go-merkle-trees/merkle"
	"github.com/ComposableFi/go-merkle-trees/mmr"
	merkletypes "github.com/ComposableFi/go-merkle-trees/types"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	scalecodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"

	// "github.com/centrifuge/go-substrate-rpc-client/v4/xxhash"

	"github.com/ethereum/go-ethereum/crypto"
	grandpa "github.com/octopus-network/ics10-grandpa-go/types"
	"github.com/octopus-network/ics10-grandpa-go/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var endpoint = "wss://rococo-rpc.polkadot.io"

func TestVerifyValidatorProof(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI(endpoint)
	if err != nil {
		// fmt.Printf("connection err,%s", err)
		log.Printf("Connecting err: %v", err)
		// t.Log("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := api.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)

	assert.NoError(t, err)
	if err != nil && err.Error() == "Method not found" {
		fmt.Printf("skipping since beefy module is not available")
		// log.Printf("skipping since beefy module is not available %v", err)
	}

	// fmt.Printf("subscribed to %s\n", polkadot_endpoint)
	log.Printf("subscribed to %s\n", endpoint)
	// assert.NoError(t, err)
	defer sub.Unsubscribe()

	timeout := time.After(24 * time.Hour)
	received := 0

	for {
		select {
		case msg := <-ch:
			log.Printf("encoded msg: %s\n", msg)

			// s := &types.SignedCommitment{}
			s := &grandpa.VersionedFinalityProof{}
			err := codec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			log.Printf("encoded msg: %#v\n", s)
			blockNumber := s.SignedCommitment.Commitment.BlockNumber
			log.Printf("blockNumber: %d\n", blockNumber)
			blockHash, err := api.RPC.Chain.GetBlockHash(uint64(blockNumber))
			require.NoError(t, err)
			log.Printf("blockHash: %#v\n", codec.HexEncodeToString(blockHash[:]))
			authorities, err := util.GetBeefyAuthorities(blockHash, api, "Authorities")
			require.NoError(t, err)
			// log.Printf("authorities: %#v\n", authorities)
			var authorityLeaves [][]byte
			for _, v := range authorities {
				authorityLeaves = append(authorityLeaves, crypto.Keccak256(v))
			}
			authorityTree, err := merkle.NewTree(hasher.Keccak256Hasher{}).FromLeaves(authorityLeaves)
			require.NoError(t, err)
			var authorityTreeRoot = grandpa.Bytes32(authorityTree.Root())
			log.Printf("authorityTreeRoot: %#v\n", codec.HexEncodeToString(authorityTreeRoot[:]))
			createBeefyAuthoritySet := grandpa.BeefyAuthoritySet{
				Id:            uint64(s.SignedCommitment.Commitment.ValidatorSetID),
				Len:           uint32(len(authorities)),
				AuthorityRoot: &authorityTreeRoot,
			}
			log.Printf("created authorityTreeRoot: %#v\n", createBeefyAuthoritySet)
			statedBeefyAuthoritySetBytes, err := util.GetBeefyAuthoritySet(blockHash, api, "BeefyAuthorities")
			require.NoError(t, err)
			log.Printf("statedBeefyAuthoritySetBytes: %#v\n", statedBeefyAuthoritySetBytes)

			csc, proofs, err := grandpa.CreateAuthorityProof(s.SignedCommitment, authorityTree)
			require.NoError(t, err)
			err = grandpa.VerifyAuthoritySignatures(csc, createBeefyAuthoritySet, proofs, authorityTreeRoot)
			require.NoError(t, err)

			received++

			if received >= 10 {
				return
			}
		case <-timeout:
			log.Printf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

func TestVerifyMmrRoot(t *testing.T) {

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

			leafIndex := uint64(clientState.GetLeafIndexForBlockNumber(blockNumber))
			log.Printf("leaf index: %d", leafIndex)
			mmrProof, err := relayApi.RPC.MMR.GenerateProof(leafIndex, blockHash)
			require.NoError(t, err)
			log.Printf("mmrProof: %#v\n", mmrProof)

			latestLeaf := mmrProof.Leaf

			BeefyNextAuthoritySetRoot := grandpa.Bytes32(latestLeaf.BeefyNextAuthoritySet.Root[:])
			parentHash := grandpa.Bytes32(latestLeaf.ParentNumberAndHash.Hash[:])

			var latestLeafMmrProof = make([][]byte, len(mmrProof.Proof.Items))
			for i := 0; i < len(mmrProof.Proof.Items); i++ {
				latestLeafMmrProof[i] = mmrProof.Proof.Items[i][:]
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
			// leafIndex := clientState.GetLeafIndexForBlockNumber(blockNumber)

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

			log.Printf("--- begin to verify relay chain mmr root ---\n")
			var signedCommitment = mmrUpdateProof.SignedCommitment
			for _, payload := range signedCommitment.Commitment.Payload {
				mmrRootID := []byte("mh")
				log.Printf("mmrRootID: %s\n", hex.EncodeToString(mmrRootID))
				log.Printf("payload.PayloadId: %s\n", hex.EncodeToString(payload.PayloadId[:]))
				// checks for the right payloadId
				if bytes.Equal(payload.PayloadId[:], mmrRootID) {
					// the next authorities are in the latest BeefyMmrLeaf

					// scale encode the mmr leaf
					mmrLeafBytes, err := scalecodec.Encode(mmrUpdateProof.MmrLeaf)
					require.NoError(t, err)
					// we treat this leaf as the latest leaf in the mmr
					mmrSize := mmr.LeafIndexToMMRSize(mmrUpdateProof.MmrLeafIndex)
					mmrLeaves := []merkletypes.Leaf{
						{
							Hash:  crypto.Keccak256(mmrLeafBytes),
							Index: mmrUpdateProof.MmrLeafIndex,
						},
					}
					mmrProof := mmr.NewProof(mmrSize, mmrUpdateProof.MmrProof, mmrLeaves, hasher.Keccak256Hasher{})

					log.Printf("expected root : %s", hex.EncodeToString(payload.PayloadData))
					root, err := mmrProof.CalculateRoot()
					require.NoError(t, err)
					log.Printf("calculated root : %s", root)

					require.Equal(t, payload.PayloadData, root)

					// verify that the leaf is valid, for the signed mmr-root-hash
					var verifyResult = mmrProof.Verify(payload.PayloadData)
					require.True(t, true, verifyResult)
					log.Printf("verify relayer mmr root successfully !\n")

					break
				}
			}

			log.Printf("Recieved Signed Commitment count: #%d", count)
		}
	}
}
