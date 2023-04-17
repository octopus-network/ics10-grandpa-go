package grandpa_test

import (
	"context"
	"sort"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/octopus-network/beefy-go/beefy"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	ibcgptypes "github.com/octopus-network/ics10-grandpa-go/grandpa"
)

func (suite *GrandpaTestSuite) TestCheckHeaderAndUpdateState() {
	suite.Suite.T().Skip("TestCheckHeaderAndUpdateState")
}

func (suite *GrandpaTestSuite) TestPruneConsensusState() {
	suite.Suite.T().Skip("TestPruneConsensusState")
}

func (suite *GrandpaTestSuite) TestSubchainLocalNet() {
	// suite.Suite.T().Skip("need to setup relaychain or subchain")
	localSubchainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localSubchainEndpoint.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	suite.Require().NoError(err)
	suite.Suite.T().Logf("subscribed to %s\n", beefy.LOCAL_RELAY_ENDPPOIT)

	defer sub.Unsubscribe()

	timeout := time.After(24 * time.Hour)
	received := 0
	// var preBlockNumber uint32
	// var preBloackHash gsrpctypes.Hash
	var clientState *ibcgptypes.ClientState
	consensusStateKVStore := make(map[uint32]ibcgptypes.ConsensusState)
	for {
		select {
		case msg := <-ch:
			suite.Suite.T().Logf("encoded msg: %s", msg)

			s := &beefy.VersionedFinalityProof{}
			err := gsrpccodec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			suite.Suite.T().Logf("decoded msg: %+v\n", s)
			// suite.Suite.T().Logf("decoded msg: %#v\n", s)
			latestSignedCommitmentBlockNumber := s.SignedCommitment.Commitment.BlockNumber
			// suite.Suite.T().Logf("blockNumber: %d\n", latestBlockNumber)
			latestSignedCommitmentBlockHash, err := localSubchainEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			suite.Require().NoError(err)
			suite.Suite.T().Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				suite.Suite.T().Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSubchainEndpoint, "BeefyAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSubchainEndpoint, "BeefyNextAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain or subchain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				suite.Suite.T().Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localSubchainEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				suite.Require().NoError(err)
				suite.Suite.T().Logf("fromBlockHash: %#x ", fromBlockHash)
				suite.Suite.T().Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)

				clientState = &ibcgptypes.ClientState{
					ChainId:               "sub-0",
					ChainType:             beefy.CHAINTYPE_SOLOCHAIN,
					ParachainId:           0,
					BeefyActivationHeight: beefy.BEEFY_ACTIVATION_BLOCK,
					LatestBeefyHeight:     clienttypes.NewHeight(clienttypes.ParseChainID("sub-0"), uint64(latestSignedCommitmentBlockNumber)),
					MmrRootHash:           s.SignedCommitment.Commitment.Payload[0].Data,
					LatestChainHeight:     clienttypes.NewHeight(clienttypes.ParseChainID("sub-0"), uint64(latestSignedCommitmentBlockNumber)),
					FrozenHeight:          clienttypes.NewHeight(clienttypes.ParseChainID("sub-0"), 0),
					AuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Len:  uint32(authoritySet.Len),
						Id:   uint64(authoritySet.ID),
						Root: authoritySet.Root[:],
					},
					NextAuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(nextAuthoritySet.ID),
						Len:  uint32(nextAuthoritySet.Len),
						Root: nextAuthoritySet.Root[:],
					},
				}
				suite.Suite.T().Logf("init client state: %+v", &clientState)
				suite.Suite.T().Logf("init client state: %+v", *clientState)
				// test pb marshal and unmarshal
				marshalCS, err := clientState.Marshal()
				suite.Require().NoError(err)
				suite.Suite.T().Logf("marshal client state: %+v", marshalCS)
				// unmarshal
				// err = clientState.Unmarshal(marshalCS)
				var unmarshalCS ibcgptypes.ClientState
				err = unmarshalCS.Unmarshal(marshalCS)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("unmarshal client state: %+v", unmarshalCS)
				continue
			}

			// step1,build authourities proof for current beefy signatures
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localSubchainEndpoint, "Authorities")
			suite.Require().NoError(err)
			bsc := beefy.ConvertCommitment(s.SignedCommitment)
			suite.Suite.T().Logf("bsc: %+v", bsc)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			suite.Require().NoError(err)

			// step2,build beefy mmr
			targetHeights := []uint32{uint32(latestSignedCommitmentBlockNumber - 1)}
			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localSolochainEndpoint, &latestSignedCommitmentBlockHash, targetHeights)
			mmrBatchProof, err := beefy.BuildMMRProofs(localSubchainEndpoint, targetHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)
			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, mmrBatchProof, authorityProof)
			suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build subchain header map
			subchainHeaderMap, err := beefy.BuildSubchainHeaderMap(localSubchainEndpoint, mmrBatchProof.Proof.LeafIndexes, "sub-0")
			suite.Require().NoError(err)
			suite.Suite.T().Logf("subchainHeaderMap: %+v", subchainHeaderMap)
			suite.Suite.T().Logf("subchainHeaderMap: %#v", subchainHeaderMap)

			pbHeader_subchainMap := ibcgptypes.ToPBSubchainHeaderMap(clientState.ChainId, subchainHeaderMap)
			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: pbBeefyMMR,
				Message:  &pbHeader_subchainMap,
			}
			suite.Suite.T().Logf("pbHeader: %+v", pbHeader)
			suite.Suite.T().Logf("pbHeader: %#v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			suite.Require().NoError(err)
			suite.Suite.T().Logf("marshalPBHeader: %+v", marshalPBHeader)

			suite.Suite.T().Log("\n\n------------- mock verify on chain -------------")
			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("unmarshal gHeader: %+v", unmarshalPBHeader)

			// step1:verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)
			suite.Suite.T().Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment
			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			suite.Suite.T().Logf("rebuildBSC: %+v", rebuildBSC)

			suite.Suite.T().Log("\n------------------ VerifySignatures --------------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifySignatures end ----------------------\n")

			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			suite.Suite.T().Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			rebuildMMRBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof)
			suite.Suite.T().Logf("Convert2MMRBatchProof: %+v", rebuildMMRBatchProof)

			suite.Suite.T().Log("\n------------------ VerifyMMRBatchProof --------------------------")
			// check mmr height
			// suite.Require().Less(clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			// result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
			// 	unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, rebuildMMRBatchProof)
			// 	suite.Require().NoError(err)
			// 	suite.Require().True(result)
			err = clientState.VerifyMMR(rebuildBSC, unmarshalBeefyMmr.MmrSize,
				rebuildMMRLeaves, rebuildMMRBatchProof)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifyMMRBatchProof end ----------------------\n")

			// step3, verify header
			// convert pb subchain header to beefy subchain header
			rebuildSubchainHeaderMap := make(map[uint32]beefy.SubchainHeader)
			unmarshalSubchainHeaderMap := unmarshalPBHeader.GetSubchainHeaderMap()
			for num, header := range unmarshalSubchainHeaderMap.GetSubchainHeaderMap() {
				rebuildSubchainHeaderMap[num] = beefy.SubchainHeader{
					BlockHeader: header.BlockHeader,
					Timestamp:   beefy.StateProof(header.Timestamp),
				}
			}
			// suite.Suite.T().Logf("rebuildSolochainHeaderMap: %+v", rebuildSolochainHeaderMap)
			suite.Suite.T().Logf("unmarshal subchainHeaderMap: %+v", *unmarshalSubchainHeaderMap)

			suite.Suite.T().Log("\n------------------ VerifySolochainHeader --------------------------")
			// err = beefy.VerifySolochainHeader(rebuildMMRLeaves, rebuildSolochainHeaderMap)
			// suite.Require().NoError(err)
			err = clientState.CheckHeader(unmarshalPBHeader, rebuildMMRLeaves)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifySolochainHeader end ----------------------\n")

			// step4, update client state
			// update client height
			clientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			clientState.LatestChainHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			clientState.MmrRootHash = unmarshalPBSC.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestNextAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range rebuildMMRLeaves {
				if latestAuthoritySetId < uint64(leaf.BeefyNextAuthoritySet.ID) {
					latestAuthoritySetId = uint64(leaf.BeefyNextAuthoritySet.ID)
					latestNextAuthoritySet = &ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
						Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
						Root: leaf.BeefyNextAuthoritySet.Root[:],
					}
				}
			}
			suite.Suite.T().Logf("current clientState.AuthoritySet.Id: %+v", clientState.AuthoritySet.Id)
			suite.Suite.T().Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			suite.Suite.T().Logf("latestNextAuthoritySet: %+v", *latestNextAuthoritySet)
			//  update client state authority set
			if clientState.AuthoritySet.Id < latestAuthoritySetId {
				clientState.AuthoritySet = *latestNextAuthoritySet
				suite.Suite.T().Logf("update clientState.AuthoritySet : %+v", clientState.AuthoritySet)
			}
			// print latest client state
			suite.Suite.T().Logf("updated client state: %+v", clientState)

			// step5,update consensue state
			var latestHeight uint32
			for _, header := range unmarshalSubchainHeaderMap.SubchainHeaderMap {
				var decodeHeader gsrpctypes.Header
				err = gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
				suite.Require().NoError(err)
				var timestamp gsrpctypes.U64
				err = gsrpccodec.Decode(header.Timestamp.Value, &timestamp)
				suite.Require().NoError(err)

				consensusState := ibcgptypes.ConsensusState{
					Timestamp: time.UnixMilli(int64(timestamp)),
					Root:      decodeHeader.StateRoot[:],
				}
				consensusStateKVStore[uint32(decodeHeader.Number)] = consensusState

				if latestHeight < uint32(decodeHeader.Number) {
					latestHeight = uint32(decodeHeader.Number)
				}
			}
			suite.Suite.T().Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)
			suite.Suite.T().Logf("latest height and consensus state: %d,%+v", latestHeight, consensusStateKVStore[latestHeight])

			// step6, mock to build and verify state proof
			for num, consnesue := range consensusStateKVStore {
				targetBlockHash, err := localSubchainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				suite.Require().NoError(err)
				timestampProof, err := beefy.GetTimestampProof(localSubchainEndpoint, targetBlockHash)
				suite.Require().NoError(err)

				proofs := make([][]byte, len(timestampProof.Proof))
				for i, proof := range timestampProof.Proof {
					proofs[i] = proof
				}
				suite.Suite.T().Logf("timestampProof proofs: %+v", proofs)
				paraTimestampStoragekey := beefy.CreateStorageKeyPrefix("Timestamp", "Now")
				suite.Suite.T().Logf("timestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				suite.Suite.T().Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				suite.Suite.T().Log("\n------------------ VerifyStateProof --------------------------")
				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
				suite.Suite.T().Log("\n------------------ VerifyStateProof end ----------------------\n")
			}

			received++
			if received >= 1 {
				return
			}
		case <-timeout:
			suite.Suite.T().Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

func (suite *GrandpaTestSuite) TestParachainLocalNet() {
	// suite.Suite.T().Skip("need setup relay chain and parachain")
	localRelayEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localRelayEndpoint.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	suite.Require().NoError(err)
	suite.Suite.T().Logf("subscribed to relaychain %s\n", beefy.LOCAL_RELAY_ENDPPOIT)
	defer sub.Unsubscribe()

	localParachainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_PARACHAIN_ENDPOINT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	suite.Suite.T().Logf("subscribed to parachain %s\n", beefy.LOCAL_PARACHAIN_ENDPOINT)

	timeout := time.After(24 * time.Hour)
	received := 0
	// var preBlockNumber uint32
	// var preBloackHash gsrpctypes.Hash
	var clientState *ibcgptypes.ClientState
	consensusStateKVStore := make(map[uint32]ibcgptypes.ConsensusState)
	for {
		select {
		case msg := <-ch:
			suite.Suite.T().Logf("encoded msg: %s", msg)

			s := &beefy.VersionedFinalityProof{}
			err := gsrpccodec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			suite.Suite.T().Logf("decoded msg: %+v\n", s)
			suite.Suite.T().Logf("decoded msg: %#v\n", s)
			latestSignedCommitmentBlockNumber := s.SignedCommitment.Commitment.BlockNumber
			// suite.Suite.T().Logf("blockNumber: %d\n", latestBlockNumber)
			latestSignedCommitmentBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			suite.Require().NoError(err)
			suite.Suite.T().Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				suite.Suite.T().Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localRelayEndpoint, "BeefyAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localRelayEndpoint, "BeefyNextAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				suite.Suite.T().Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				suite.Require().NoError(err)
				suite.Suite.T().Logf("fromBlockHash: %#x ", fromBlockHash)
				suite.Suite.T().Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)
				changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("changeSet len: %d", len(changeSets))
				var latestChainHeight uint32
				if len(changeSets) == 0 {
					latestChainHeight = 0
				} else {
					var packedParachainHeights []int
					for _, changeSet := range changeSets {
						for _, change := range changeSet.Changes {
							suite.Suite.T().Logf("change.StorageKey: %#x", change.StorageKey)
							suite.Suite.T().Log("change.HasStorageData: ", change.HasStorageData)
							suite.Suite.T().Logf("change.HasStorageData: %#x", change.StorageData)
							if change.HasStorageData {
								var parachainheader gsrpctypes.Header
								// first decode byte array
								var bz []byte
								err = gsrpccodec.Decode(change.StorageData, &bz)
								suite.Require().NoError(err)
								// second decode header
								err = gsrpccodec.Decode(bz, &parachainheader)
								suite.Require().NoError(err)
								packedParachainHeights = append(packedParachainHeights, int(parachainheader.Number))
							}
						}
					}
					suite.Suite.T().Logf("raw packedParachainHeights: %+v", packedParachainHeights)
					// sort heights and find latest height
					sort.Sort(sort.Reverse(sort.IntSlice(packedParachainHeights)))
					suite.Suite.T().Logf("sort.Reverse: %+v", packedParachainHeights)
					latestChainHeight = uint32(packedParachainHeights[0])
					suite.Suite.T().Logf("latestHeight: %d", latestChainHeight)
				}

				clientState = &ibcgptypes.ClientState{
					ChainId:               "parachain-1",
					ChainType:             beefy.CHAINTYPE_PARACHAIN,
					ParachainId:           beefy.LOCAL_PARACHAIN_ID,
					BeefyActivationHeight: beefy.BEEFY_ACTIVATION_BLOCK,
					LatestBeefyHeight:     clienttypes.NewHeight(clienttypes.ParseChainID("parachain-1"), uint64(latestSignedCommitmentBlockNumber)),
					MmrRootHash:           s.SignedCommitment.Commitment.Payload[0].Data,
					LatestChainHeight:     clienttypes.NewHeight(clienttypes.ParseChainID("parachain-1"), uint64(latestChainHeight)),
					FrozenHeight:          clienttypes.NewHeight(clienttypes.ParseChainID("parachain-1"), 0),
					AuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(authoritySet.ID),
						Len:  uint32(authoritySet.Len),
						Root: authoritySet.Root[:],
					},
					NextAuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(nextAuthoritySet.ID),
						Len:  uint32(nextAuthoritySet.Len),
						Root: nextAuthoritySet.Root[:],
					},
				}
				suite.Suite.T().Logf("init client state: %+v", &clientState)
				suite.Suite.T().Logf("init client state: %+v", *clientState)
				// test pb marshal and unmarshal
				marshalCS, err := clientState.Marshal()
				suite.Require().NoError(err)
				suite.Suite.T().Logf("marshal client state: %+v", marshalCS)
				// unmarshal
				// err = clientState.Unmarshal(marshalCS)
				var unmarshalCS ibcgptypes.ClientState
				err = unmarshalCS.Unmarshal(marshalCS)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("unmarshal client state: %+v", unmarshalCS)
				continue
			}

			// step1,build authourities proof for current beefy signatures
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localRelayEndpoint, "Authorities")
			suite.Require().NoError(err)
			bsc := beefy.ConvertCommitment(s.SignedCommitment)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			suite.Require().NoError(err)

			// step2,build beefy mmr
			fromBlockNumber := clientState.LatestBeefyHeight.RevisionHeight + 1

			fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
			suite.Require().NoError(err)

			changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
			suite.Require().NoError(err)

			var targetRelaychainBlockHeights []uint32
			for _, changeSet := range changeSets {
				relayHeader, err := localRelayEndpoint.RPC.Chain.GetHeader(changeSet.Block)
				suite.Require().NoError(err)
				targetRelaychainBlockHeights = append(targetRelaychainBlockHeights, uint32(relayHeader.Number))
			}

			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localRelayEndpoint, &latestSignedCommitmentBlockHash, targetRelaychainBlockHeights)
			mmrBatchProof, err := beefy.BuildMMRProofs(localRelayEndpoint, targetRelaychainBlockHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)

			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, mmrBatchProof, authorityProof)
			suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build parachain header proof and verify that proof
			parachainHeaderMap, err := beefy.BuildParachainHeaderMap(localRelayEndpoint, localParachainEndpoint,
				mmrBatchProof.Proof.LeafIndexes, "parachain-1", beefy.LOCAL_PARACHAIN_ID)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("parachainHeaderMap: %+v", parachainHeaderMap)

			// convert beefy parachain header to pb parachain header
			pbHeader_parachainMap := ibcgptypes.ToPBParachainHeaderMap(clientState.ChainId, parachainHeaderMap)
			suite.Suite.T().Logf("pbHeader_parachainMap: %+v", pbHeader_parachainMap)

			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: pbBeefyMMR,
				Message:  &pbHeader_parachainMap,
			}
			suite.Suite.T().Logf("gpheader: %+v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			suite.Require().NoError(err)
			// suite.Suite.T().Logf("marshal gHeader: %+v", marshalGHeader)

			suite.Suite.T().Log("\n------------- mock verify on chain -------------\n")

			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("unmarshal gHeader: %+v", unmarshalPBHeader)

			// verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// suite.Suite.T().Logf("gBeefyMMR: %+v", gBeefyMMR)
			suite.Suite.T().Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment

			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			suite.Suite.T().Logf("rebuildBSC: %+v", rebuildBSC)

			suite.Suite.T().Log("------------------ verify signature ----------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			suite.Require().NoError(err)
			suite.Suite.T().Log("--------------verify signature end ----------------------------")

			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			suite.Suite.T().Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			beefyMmrBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof)
			suite.Suite.T().Logf("Convert2MMRBatchProof: %+v", beefyMmrBatchProof)

			suite.Suite.T().Log("\n---------------- verify mmr proof --------------------")
			// check mmr height
			// suite.Require().Less(clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			// result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
			// suite.Require().True(result)
			// 	unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, beefyMmrBatchProof)
			err = clientState.VerifyMMR(rebuildBSC, unmarshalBeefyMmr.MmrSize,
				rebuildMMRLeaves, beefyMmrBatchProof)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n-------------verify mmr proof end--------------------\n")

			// step3, verify header
			// convert pb parachain header to beefy parachain header
			rebuildParachainHeaderMap := make(map[uint32]beefy.ParachainHeader)
			unmarshalParachainHeaderMap := unmarshalPBHeader.GetParachainHeaderMap()
			for num, header := range unmarshalParachainHeaderMap.ParachainHeaderMap {
				rebuildParachainHeaderMap[num] = beefy.ParachainHeader{
					ParaId:      header.ParachainId,
					BlockHeader: header.BlockHeader,
					Proof:       header.Proofs,
					HeaderIndex: header.HeaderIndex,
					HeaderCount: header.HeaderCount,
					Timestamp:   beefy.StateProof(header.Timestamp),
				}
			}
			// suite.Suite.T().Logf("parachainHeaderMap: %+v", parachainHeaderMap)
			// suite.Suite.T().Logf("gParachainHeaderMap: %+v", gParachainHeaderMap)
			suite.Suite.T().Logf("unmarshal parachainHeaderMap: %+v", *unmarshalParachainHeaderMap)
			// suite.Suite.T().Logf("rebuildSolochainHeaderMap: %+v", rebuildParachainHeaderMap)
			suite.Suite.T().Log("\n----------- VerifyParachainHeader -----------")
			// err = beefy.VerifyParachainHeader(rebuildMMRLeaves, rebuildParachainHeaderMap)
			// suite.Require().NoError(err)
			err = clientState.CheckHeader(unmarshalPBHeader, rebuildMMRLeaves)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n--------------------------------------------\n")

			// step4, update client state
			clientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			clientState.MmrRootHash = unmarshalPBSC.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestNextAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range rebuildMMRLeaves {
				if latestAuthoritySetId < uint64(leaf.BeefyNextAuthoritySet.ID) {
					latestAuthoritySetId = uint64(leaf.BeefyNextAuthoritySet.ID)
					latestNextAuthoritySet = &ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
						Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
						Root: leaf.BeefyNextAuthoritySet.Root[:],
					}
				}
			}
			suite.Suite.T().Logf("current clientState.AuthoritySet.Id: %+v", clientState.AuthoritySet.Id)
			suite.Suite.T().Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			suite.Suite.T().Logf("latestNextAuthoritySet: %+v", *latestNextAuthoritySet)
			//  update client state authority set
			if clientState.AuthoritySet.Id < latestAuthoritySetId {
				clientState.AuthoritySet = *latestNextAuthoritySet
				suite.Suite.T().Logf("update clientState.AuthoritySet : %+v", clientState.AuthoritySet)
			}

			// print latest client state
			suite.Suite.T().Logf("updated client state: %+v", clientState)

			// step5,update consensue state
			var latestHeight uint32
			for _, header := range unmarshalParachainHeaderMap.ParachainHeaderMap {
				var decodeHeader gsrpctypes.Header
				err = gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
				suite.Require().NoError(err)
				var timestamp gsrpctypes.U64
				err = gsrpccodec.Decode(header.Timestamp.Value, &timestamp)
				suite.Require().NoError(err)

				consensusState := ibcgptypes.ConsensusState{
					Timestamp: time.UnixMilli(int64(timestamp)),
					Root:      decodeHeader.StateRoot[:],
				}
				// note: the block number must be parachain header blocknumber,not relaychain header block number
				consensusStateKVStore[uint32(decodeHeader.Number)] = consensusState
				if latestHeight < uint32(decodeHeader.Number) {
					latestHeight = uint32(decodeHeader.Number)
				}
			}
			suite.Suite.T().Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)
			suite.Suite.T().Logf("latest height and consensus state: %d,%+v", latestHeight, consensusStateKVStore[latestHeight])

			// update client state latest chain height
			clientState.LatestChainHeight = clienttypes.NewHeight(clienttypes.ParseChainID(clientState.ChainId), uint64(latestHeight))

			// step6, mock to build and verify state proof
			for num, consnesue := range consensusStateKVStore {
				// Note: get data from parachain
				targetBlockHash, err := localParachainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				suite.Require().NoError(err)
				timestampProof, err := beefy.GetTimestampProof(localParachainEndpoint, targetBlockHash)
				suite.Require().NoError(err)

				proofs := make([][]byte, len(timestampProof.Proof))
				for i, proof := range timestampProof.Proof {
					proofs[i] = proof
				}
				suite.Suite.T().Logf("timestampProof proofs: %+v", proofs)
				paraTimestampStoragekey := beefy.CreateStorageKeyPrefix("Timestamp", "Now")
				suite.Suite.T().Logf("paraTimestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				suite.Suite.T().Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				suite.Suite.T().Log("\n------------- mock verify state proof-------------")

				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
				suite.Suite.T().Log("\n--------------------------------------------------")
			}

			received++
			if received >= 3 {
				return
			}
		case <-timeout:
			suite.Suite.T().Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}
