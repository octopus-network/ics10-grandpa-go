package grandpa

import (
	"errors"
	"time"

	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/octopus-network/beefy-go/beefy"
)

// VerifyClientMessage checks if the clientMessage is of type Header or Misbehaviour and verifies the message
func (cs *ClientState) VerifyClientMessage(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	clientMsg exported.ClientMessage,
) error {
	panic("")
}

// step1: verify signature
// step2: verify mmr
// step3: verfiy header
// step4: update client state and consensuse state
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	header exported.ClientMessage,
) (exported.ClientState, exported.ConsensusState, error) {
	ctx.Logger().Debug("LightClient:", "10-Grandpa", "method:", "CheckHeaderAndUpdateState")

	pbHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", &Header{}, header,
		)
	}

	beefyMMR := pbHeader.BeefyMmr

	// step1:  verify signature
	// convert signedcommitment
	bsc := ToBeefySC(beefyMMR.SignedCommitment)
	err := cs.VerifySignatures(bsc, beefyMMR.SignatureProofs)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to verify signatures")
	}

	// step2: verify mmr
	// convert mmrleaf
	beefyMMRLeaves := ToBeefyMMRLeaves(beefyMMR.MmrLeavesAndBatchProof.Leaves)
	// convert mmr proof
	beefyBatchProof := ToMMRBatchProof(beefyMMR.MmrLeavesAndBatchProof)
	// verify mmr
	err = cs.VerifyMMR(bsc, beefyMMR.MmrSize, beefyMMRLeaves, beefyBatchProof)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to verify mmr")
	}

	// step3: verify header
	err = cs.VerifyHeader(*pbHeader, beefyMMRLeaves)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to verify header")
	}

	// step4: update state
	// update client state and build new client state
	newClientState, err := cs.UpdateClientState(ctx, beefyMMR.SignedCommitment.Commitment, beefyMMR.MmrLeavesAndBatchProof.Leaves)
	if err != nil {
		return nil, nil, err
	}
	// update consensue state and build latest new consensue state
	newConsensusState, err := cs.UpdateConsensusStates(ctx, cdc, clientStore, pbHeader)
	if err != nil {
		return nil, nil, err
	}
	return newClientState, newConsensusState, nil
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
	beefyMMRLeaves []gsrpctypes.MMRLeaf, mmrBatchProof beefy.MMRBatchProof) error {
	// check mmr height
	if bsc.Commitment.BlockNumber <= cs.LatestBeefyHeight {
		return sdkerrors.Wrap(errors.New("Commitment.BlockNumber < cs.LatestBeefyHeight"), "")
	}

	//verify mmr proof
	result, err := beefy.VerifyMMRBatchProof(bsc.Commitment.Payload, mmrSize, beefyMMRLeaves, mmrBatchProof)
	if err != nil || !result {
		return sdkerrors.Wrap(errors.New("failed to verify mmr proof"), "")
	}
	return nil
}

// verify solochain header or parachain header
func (cs ClientState) VerifyHeader(gpHeader Header, beefyMMRLeaves []gsrpctypes.MMRLeaf,
) error {

	switch cs.ChainType {
	case beefy.CHAINTYPE_SOLOCHAIN:
		Logger.Debug("verify solochain header")

		headerMap := gpHeader.GetSubchainHeaderMap()
		// convert pb solochain header to beefy solochain header
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
		Logger.Debug("verify parachain header")

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

// update client state
func (cs ClientState) UpdateClientState(ctx sdk.Context, commitment Commitment, mmrLeaves []MMRLeaf) (*ClientState, error) {
	ctx.Logger().Debug("update client state")
	// var newClientState *ClientState

	latestBeefyHeight := commitment.BlockNumber
	latestChainHeight := latestBeefyHeight
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
		cs.ParachainId, cs.BeefyActivationBlock,
		latestBeefyHeight, mmrRoot, latestChainHeight,
		cs.FrozenHeight, *latestNextAuthoritySet, cs.NextAuthoritySet)

	return newClientState, nil

}

// TODO: save all the consensue state at height,but just return latest block header
func (cs ClientState) UpdateConsensusStates(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, header *Header) (*ConsensusState, error) {

	var newConsensueState *ConsensusState
	var latestChainHeight uint32
	var latestBlockHeader gsrpctypes.Header
	var latestTimestamp uint64
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
			err = updateConsensuestate(clientStore, cdc, decodeHeader, uint64(decodeTimestamp))
			if err != nil {
				return nil, sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "update consensus state failed")
			}
			//find latest header and timestmap
			if latestChainHeight < uint32(decodeHeader.Number) {
				latestChainHeight = uint32(decodeHeader.Number)
				latestBlockHeader = decodeHeader
				latestTimestamp = uint64(decodeTimestamp)
			}
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
			err = updateConsensuestate(clientStore, cdc, decodeHeader, uint64(decodeTimestamp))
			if err != nil {
				return nil, sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "update consensus state failed")
			}
			//find latest header and timestmap
			if latestChainHeight < uint32(decodeHeader.Number) {
				latestChainHeight = uint32(decodeHeader.Number)
				latestBlockHeader = decodeHeader
				latestTimestamp = uint64(decodeTimestamp)
			}

		}
	}

	// TODO: asset blockheader and timestamp is not nil
	newConsensueState = NewConsensusState(latestBlockHeader.StateRoot[:], time.UnixMilli(int64(latestTimestamp)))

	// set metadata for this consensus
	latestHeigh := clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: uint64(latestChainHeight),
	}
	setConsensusMetadata(ctx, clientStore, latestHeigh)
	// TODO: consider to pruning!
	//
	return newConsensueState, nil

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
	// header, ok := clientMsg.(*Header)
	panic("")
}

// UpdateStateOnMisbehaviour updates state upon misbehaviour, freezing the ClientState. This method should only be called when misbehaviour is detected
// as it does not perform any misbehaviour checks.
func (cs ClientState) UpdateStateOnMisbehaviour(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, _ exported.ClientMessage) {
	panic("")
}
