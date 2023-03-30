package grandpa

import (
	"errors"
	fmt "fmt"

	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/octopus-network/beefy-go/beefy"
)

// VerifyClientMessage checks if the clientMessage is of type Header or Misbehaviour and verifies the message
func (cs *ClientState) VerifyClientMessage(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	clientMsg exported.ClientMessage,
) error {
	switch msg := clientMsg.(type) {
	case *Header:
		return cs.verifyHeader(ctx, clientStore, cdc, msg)

	case *Misbehaviour:
		return cs.verifyMisbehaviour(ctx, clientStore, cdc, msg)
	default:
		return clienttypes.ErrInvalidClientType
	}
}

// step1: verify signature
// step2: verify mmr
// step3: verfiy header
func (cs ClientState) verifyHeader(
	ctx sdk.Context, clientStore sdk.KVStore, cdc codec.BinaryCodec,
	header *Header,
) error {
	ctx.Logger().Debug("LightClient:", "10-Grandpa", "method:", "verifyHeader")

	beefyMMR := header.BeefyMmr

	// step1:  verify signature
	// convert signedcommitment
	bsc := ToBeefySC(beefyMMR.SignedCommitment)
	err := cs.VerifySignatures(bsc, beefyMMR.SignatureProofs)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to verify signatures")
	}

	// step2: verify mmr
	// convert mmrleaf
	beefyMMRLeaves := ToBeefyMMRLeaves(beefyMMR.MmrLeavesAndBatchProof.Leaves)
	// convert mmr proof
	beefyBatchProof := ToMMRBatchProof(beefyMMR.MmrLeavesAndBatchProof)
	// verify mmr
	err = cs.VerifyMMR(bsc, beefyMMR.MmrSize, beefyMMRLeaves, beefyBatchProof)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to verify mmr")
	}

	// step3: verify header
	err = cs.CheckHeader(*header, beefyMMRLeaves)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to verify header")
	}

	return nil
}

// verify signatures
func (cs ClientState) VerifySignatures(bsc beefy.SignedCommitment, SignatureProofs [][]byte) error {
	// checking signatures Threshold
	if beefy.SignatureThreshold(cs.AuthoritySet.Len) > uint32(len(bsc.Signatures)) ||
		beefy.SignatureThreshold(cs.NextAuthoritySet.Len) > uint32(len(bsc.Signatures)) {

		return sdkerrors.Wrap(errors.New("verify signature error "), ErrInvalidValidatorSet.Error())
	}
	// verify signatures
	switch bsc.Commitment.ValidatorSetID {
	case cs.AuthoritySet.Id:
		err := beefy.VerifySignature(bsc, uint64(cs.AuthoritySet.Len), beefy.Bytes32(cs.AuthoritySet.Root), SignatureProofs)
		if err != nil {
			return sdkerrors.Wrap(err, ErrInvalidValidatorSet.Error())
		}

	case cs.NextAuthoritySet.Id:
		err := beefy.VerifySignature(bsc, uint64(cs.NextAuthoritySet.Len), beefy.Bytes32(cs.NextAuthoritySet.Root), SignatureProofs)
		if err != nil {
			return sdkerrors.Wrap(err, ErrInvalidValidatorSet.Error())
		}
	}

	return nil
}

// verify batch mmr proof
func (cs ClientState) VerifyMMR(bsc beefy.SignedCommitment, mmrSize uint64,
	beefyMMRLeaves []gsrpctypes.MMRLeaf, mmrBatchProof beefy.MMRBatchProof,
) error {
	// check mmr height
	if bsc.Commitment.BlockNumber <= uint32(cs.LatestBeefyHeight.RevisionHeight) {
		return sdkerrors.Wrap(errors.New("Commitment.BlockNumber < cs.LatestBeefyHeight"), "")
	}

	// verify mmr proof
	result, err := beefy.VerifyMMRBatchProof(bsc.Commitment.Payload, mmrSize, beefyMMRLeaves, mmrBatchProof)
	if err != nil || !result {
		return sdkerrors.Wrap(errors.New("failed to verify mmr proof"), "")
	}
	return nil
}

// verify solochain header or parachain header
func (cs ClientState) CheckHeader(gpHeader Header, beefyMMRLeaves []gsrpctypes.MMRLeaf,
) error {
	switch cs.ChainType {
	case beefy.CHAINTYPE_SOLOCHAIN:
		logger.Sugar().Debug("verify solochain header")

		headerMap := gpHeader.GetSubchainHeaderMap()
		// convert pb solochain header to beefy subchain header
		beefySubchainHeaderMap := make(map[uint32]beefy.SubchainHeader)
		for num, header := range headerMap.SubchainHeaderMap {
			beefySubchainHeaderMap[num] = beefy.SubchainHeader{
				BlockHeader: header.BlockHeader,
				Timestamp:   beefy.StateProof(header.Timestamp),
			}
		}
		err := beefy.VerifySubchainHeader(beefyMMRLeaves, beefySubchainHeaderMap)
		if err != nil {
			return err
		}

	case beefy.CHAINTYPE_PARACHAIN:
		logger.Sugar().Debug("verify parachain header")

		headerMap := gpHeader.GetParachainHeaderMap()
		// convert pb parachain header to beefy parachain header
		beefyParachainHeaderMap := make(map[uint32]beefy.ParachainHeader)
		for num, header := range headerMap.ParachainHeaderMap {
			beefyParachainHeaderMap[num] = beefy.ParachainHeader{
				ParaId:      header.ParachainId,
				BlockHeader: header.BlockHeader,
				Proof:       header.Proofs,
				HeaderIndex: header.HeaderIndex,
				HeaderCount: header.HeaderCount,
				Timestamp:   beefy.StateProof(header.Timestamp),
			}
		}
		err := beefy.VerifyParachainHeader(beefyMMRLeaves, beefyParachainHeaderMap)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateState may be used to either create a consensus state for:
// - a future height greater than the latest client state height
// - a past height that was skipped during bisection
// If we are updating to a past height, a consensus state is created for that height to be persisted in client store
// If we are updating to a future height, the consensus state is created and the client state is updated to reflect
// the new latest height
// A list containing the updated consensus height is returned.
// UpdateState must only be used to update within a single revision, thus header revision number and trusted height's revision
// number must be the same. To update to a new revision, use a separate upgrade path
// UpdateState will prune the oldest consensus state if it is expired.
func (cs ClientState) UpdateState(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, clientMsg exported.ClientMessage) []exported.Height {
	header, ok := clientMsg.(*Header)
	if !ok {
		panic(fmt.Errorf("expected type %T, got %T", &Header{}, clientMsg))
	}
	chainID, latestHeader, _, err := getLastestChainHeader(header)

	beefyMMR := header.BeefyMmr
	commitment := beefyMMR.SignedCommitment.Commitment
	mmrLeaves := beefyMMR.MmrLeavesAndBatchProof.Leaves
	// var newClientState *ClientState
	latestBeefyHeight := clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(commitment.BlockNumber))
	// latestBeefyHeight := commitment.BlockNumber
	latestChainHeight := clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestHeader.Number))
	mmrRoot := commitment.Payloads[0].Data

	// find latest next authority set from mmrleaves
	var latestNextAuthoritySet *BeefyAuthoritySet
	var latestAuthoritySetId uint64
	for _, leaf := range mmrLeaves {
		if latestAuthoritySetId < leaf.BeefyNextAuthoritySet.Id {
			latestAuthoritySetId = leaf.BeefyNextAuthoritySet.Id
			latestNextAuthoritySet = &BeefyAuthoritySet{
				Id:   leaf.BeefyNextAuthoritySet.Id,
				Len:  leaf.BeefyNextAuthoritySet.Len,
				Root: leaf.BeefyNextAuthoritySet.Root,
			}
		}
	}

	newClientState := NewClientState(cs.ChainType, cs.ChainId,
		cs.ParachainId, cs.BeefyActivationHeight,
		latestBeefyHeight, mmrRoot, latestChainHeight,
		cs.FrozenHeight, *latestNextAuthoritySet, cs.NextAuthoritySet)
	// save client state
	setClientState(clientStore, cdc, newClientState)

	// update consensue state and return all the updated heights
	updatedHeights, err := cs.UpdateConsensusStates(ctx, cdc, clientStore, header)
	if err != nil {
		return []exported.Height{}
	}

	return updatedHeights
}

// TODO: save all the consensue state at height,but just return latest block header
func (cs ClientState) UpdateConsensusStates(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, header *Header) ([]exported.Height, error) {
	// var newConsensueState *ConsensusState
	var latestChainid string
	var latestChainHeight uint32
	// var latestBlockHeader gsrpctypes.Header
	// var latestTimestamp uint64
	var updatedHeights []exported.Height
	switch cs.ChainType {
	case beefy.CHAINTYPE_SOLOCHAIN:
		subchainHeaderMap := header.GetSubchainHeaderMap().SubchainHeaderMap
		for _, header := range subchainHeaderMap {
			var decodeHeader gsrpctypes.Header
			err := gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode header error")
			}
			var decodeTimestamp gsrpctypes.U64
			err = gsrpccodec.Decode(header.Timestamp.Value, &decodeTimestamp)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode timestamp error")
			}
			height := clienttypes.NewHeight(clienttypes.ParseChainID(header.ChainId), uint64(decodeHeader.Number))
			err = updateConsensuestate(clientStore, cdc, height, decodeHeader, uint64(decodeTimestamp))
			if err != nil {
				return nil, sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "update consensus state failed")
			}
			// find latest header and timestamp
			if latestChainHeight < uint32(decodeHeader.Number) {
				latestChainHeight = uint32(decodeHeader.Number)
				latestChainid = header.ChainId
				// latestBlockHeader = decodeHeader
				// latestTimestamp = uint64(decodeTimestamp)
			}
			updatedHeights = append(updatedHeights, clienttypes.NewHeight(clienttypes.ParseChainID(header.ChainId), uint64(decodeHeader.Number)))
		}

	case beefy.CHAINTYPE_PARACHAIN:
		parachainHeaderMap := header.GetParachainHeaderMap().ParachainHeaderMap
		for _, header := range parachainHeaderMap {
			var decodeHeader gsrpctypes.Header
			err := gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode header error")
			}
			var decodeTimestamp gsrpctypes.U64
			err = gsrpccodec.Decode(header.Timestamp.Value, &decodeTimestamp)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode timestamp error")
			}
			height := clienttypes.NewHeight(clienttypes.ParseChainID(header.ChainId), uint64(decodeHeader.Number))
			err = updateConsensuestate(clientStore, cdc, height, decodeHeader, uint64(decodeTimestamp))

			if err != nil {
				return nil, sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "update consensus state failed")
			}
			// find latest header and timestamp
			if latestChainHeight < uint32(decodeHeader.Number) {
				latestChainHeight = uint32(decodeHeader.Number)
				latestChainid = header.ChainId
				// latestBlockHeader = decodeHeader
				// latestTimestamp = uint64(decodeTimestamp)
			}
			updatedHeights = append(updatedHeights, clienttypes.NewHeight(clienttypes.ParseChainID(header.ChainId), uint64(decodeHeader.Number)))
		}
	}

	// set metadata for this consensus
	latestHeigh := clienttypes.NewHeight(clienttypes.ParseChainID(latestChainid), uint64(latestChainHeight))
	setConsensusMetadata(ctx, clientStore, latestHeigh)
	// TODO: consider to pruning!
	//
	return updatedHeights, nil
}

// UpdateStateOnMisbehaviour updates state upon misbehaviour, freezing the ClientState. This method should only be called when misbehaviour is detected
// as it does not perform any misbehaviour checks.
func (cs ClientState) UpdateStateOnMisbehaviour(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, _ exported.ClientMessage) {
	cs.FrozenHeight = FrozenHeight

	clientStore.Set(host.ClientStateKey(), clienttypes.MustMarshalClientState(cdc, &cs))
}
