package grandpa

import (
	"errors"
	"time"

	ics23 "github.com/confio/ics23/go"
	"github.com/octopus-network/beefy-go/beefy"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	trieproof "github.com/octopus-network/trie-go/trie/proof"
)

var _ exported.ClientState = (*ClientState)(nil)

// NewClientState creates a new ClientState instance
func NewClientState(
	chainType uint32,
	chainId string,
	parachainId uint32,
	beefyActivationHeight uint32,
	latestBeefyHeight clienttypes.Height,
	mmrRootHash []byte,
	latestHeight clienttypes.Height,
	frozenHeight clienttypes.Height,
	authoritySet BeefyAuthoritySet,
	nextAuthoritySet BeefyAuthoritySet,
) *ClientState {
	return &ClientState{
		ChainType:             chainType,
		ChainId:               chainId,
		ParachainId:           parachainId,
		BeefyActivationHeight: beefyActivationHeight,
		LatestBeefyHeight:     latestBeefyHeight,
		MmrRootHash:           mmrRootHash,
		LatestChainHeight:     latestHeight,
		FrozenHeight:          frozenHeight,
		AuthoritySet:          authoritySet,
		NextAuthoritySet:      nextAuthoritySet,
	}
}

// GetChainID returns the chain-id
func (cs ClientState) GetChainID() string {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.GetChainID()")
	return cs.ChainId
}

// ClientType is tendermint.
func (cs ClientState) ClientType() string {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.ClientType()")
	return ModuleName
}

// TODO: which height? latest beefy height or latest chain height.
func (cs ClientState) GetLatestHeight() exported.Height {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.GetLatestHeight()")

	return cs.LatestChainHeight
}

// GetTimestampAtHeight returns the timestamp in nanoseconds of the consensus state at the given height.
func (cs ClientState) GetTimestampAtHeight(
	ctx sdk.Context,
	clientStore sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
) (uint64, error) {
	// get consensus state at height from clientStore to check for expiry
	consState, err := GetConsensusState(clientStore, cdc, height)
	if err != nil {
		return 0, err
	}
	return consState.GetTimestamp(), nil
}

// Status returns the status of the Grandpa client.
// The client may be:
// - Active: FrozenHeight is zero and client is not expired
// - Frozen: Frozen Height is not zero
// - Expired: the latest consensus state timestamp + trusting period <= current time
//
// A frozen client will become expired, so the Frozen status
// has higher precedence.
func (cs ClientState) Status(
	ctx sdk.Context,
	clientStore sdk.KVStore,
	cdc codec.BinaryCodec,
) exported.Status {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.Status()")

	if !cs.FrozenHeight.IsZero() {
		return exported.Frozen
	}

	return exported.Active
}

// TODO: check expired
// IsExpired returns whether or not the client has passed the trusting period since the last
// update (in which case no headers are considered valid).
func (cs ClientState) IsExpired(latestTimestamp, now time.Time) bool {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.IsExpired()")
	return false
}

// Validate performs a basic validation of the client state fields.
func (cs ClientState) Validate() error {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.Validate()")

	if cs.LatestBeefyHeight.RevisionHeight == 0 {
		return sdkerrors.Wrap(ErrInvalidHeaderHeight, "beefy height cannot be zero")
	}

	return nil
}

// GetProofSpecs returns the format the client expects for proof verification
// as a string array specifying the proof type for each position in chained proof
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.GetProofSpecs()")

	ps := []*ics23.ProofSpec{}
	// ps := commitmenttypes.GetSDKSpecs()
	return ps
}

// ZeroCustomFields returns a ClientState that is a copy of the current ClientState
// with all client customizable fields zeroed out
func (cs ClientState) ZeroCustomFields() exported.ClientState {
	logger.Sugar().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.ZeroCustomFields()")

	// copy over all chain-specified fields
	// and leave custom fields empty

	return &ClientState{
		ChainId:               cs.ChainId,
		ChainType:             cs.ChainType,
		BeefyActivationHeight: cs.BeefyActivationHeight,
		LatestBeefyHeight:     cs.LatestBeefyHeight,
		MmrRootHash:           cs.MmrRootHash,
		LatestChainHeight:     cs.LatestChainHeight,
		FrozenHeight:          cs.FrozenHeight,
		AuthoritySet:          cs.AuthoritySet,
		NextAuthoritySet:      cs.NextAuthoritySet,
	}
}

// Initialize will check that initial consensus state is a Grandpa consensus state
// and will store ProcessedTime for initial consensus state as ctx.BlockTime()
func (cs ClientState) Initialize(ctx sdk.Context, _ codec.BinaryCodec, clientStore sdk.KVStore, consState exported.ConsensusState) error {
	ctx.Logger().Debug("LightClient:", "10-Grandpa", "method:", "ClientState.Initialize()")

	if _, ok := consState.(*ConsensusState); !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "invalid initial consensus state. expected type: %T, got: %T",
			&ConsensusState{}, consState)
	}
	// set metadata for initial consensus state.
	// Note,this height must be solochain or parachain height
	latestChainHeigh := cs.GetLatestHeight()
	setConsensusMetadata(ctx, clientStore, latestChainHeigh)

	return nil
}

// VerifyMembership is a generic proof verification method which verifies a proof of the existence of a value at a given CommitmentPath at the specified height.
// The caller is expected to construct the full CommitmentPath from a CommitmentPrefix and a standardized path (as defined in ICS 24).
// If a zero proof height is passed in, it will fail to retrieve the associated consensus state.
func (cs ClientState) VerifyMembership(
	ctx sdk.Context,
	clientStore sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	delayTimePeriod uint64,
	delayBlockPeriod uint64,
	proof []byte,
	path exported.Path,
	value []byte,
) error {
	if cs.GetLatestHeight().LT(height) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state height < proof height (%d < %d), please ensure the client has been updated", cs.GetLatestHeight(), height,
		)
	}
	if err := verifyDelayPeriodPassed(ctx, clientStore, height, delayTimePeriod, delayBlockPeriod); err != nil {
		return err
	}

	// var merkleProof commitmenttypes.MerkleProof
	// if err := cdc.Unmarshal(proof, &merkleProof); err != nil {
	// 	return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "failed to unmarshal proof into ICS 23 commitment merkle proof")
	// }

	// merklePath, ok := path.(commitmenttypes.MerklePath)
	// if !ok {
	// 	return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", commitmenttypes.MerklePath{}, path)
	// }

	// consensusState, found := GetConsensusState(clientStore, cdc, height)
	// if !found {
	// 	return sdkerrors.Wrap(clienttypes.ErrConsensusStateNotFound, "please ensure the proof was constructed against a height that exists on the client")
	// }

	// if err := merkleProof.VerifyMembership(cs.ProofSpecs, consensusState.GetRoot(), merklePath, value); err != nil {
	// 	return err
	// }

	// Note: decode proof
	// err = gsrpccodec.Decode(proof, &stateProof)
	var stateProof StateProof
	err := cdc.Unmarshal(proof, &stateProof)
	if err != nil {
		return sdkerrors.Wrap(err, "proof couldn't be decoded into StateProof struct")
	}

	consensusState, err := GetConsensusState(clientStore, cdc, height)
	if err != nil {
		return err
	}

	// TODO: check value codec?
	// err = beefy.VerifyStateProof(stateProof.Proofs, provingConsensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, value)
	if err != nil {
		logger.Sugar().Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyClientState()", "failure to verify state proof: ", err)
		return err
	}

	return nil
}

// VerifyNonMembership is a generic proof verification method which verifies the absence of a given CommitmentPath at a specified height.
// The caller is expected to construct the full CommitmentPath from a CommitmentPrefix and a standardized path (as defined in ICS 24).
// If a zero proof height is passed in, it will fail to retrieve the associated consensus state.
func (cs ClientState) VerifyNonMembership(
	ctx sdk.Context,
	clientStore sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	delayTimePeriod uint64,
	delayBlockPeriod uint64,
	proof []byte,
	path exported.Path,
) error {
	if cs.GetLatestHeight().LT(height) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state height < proof height (%d < %d), please ensure the client has been updated", cs.GetLatestHeight(), height,
		)
	}

	if err := verifyDelayPeriodPassed(ctx, clientStore, height, delayTimePeriod, delayBlockPeriod); err != nil {
		return err
	}

	var merkleProof commitmenttypes.MerkleProof
	if err := cdc.Unmarshal(proof, &merkleProof); err != nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "failed to unmarshal proof into ICS 23 commitment merkle proof")
	}

	// merklePath, ok := path.(commitmenttypes.MerklePath)
	// if !ok {
	// 	return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", commitmenttypes.MerklePath{}, path)
	// }

	// consensusState, found := GetConsensusState(clientStore, cdc, height)
	// if !found {
	// 	return sdkerrors.Wrap(clienttypes.ErrConsensusStateNotFound, "please ensure the proof was constructed against a height that exists on the client")
	// }

	// if err := merkleProof.VerifyNonMembership(cs.ProofSpecs, consensusState.GetRoot(), merklePath); err != nil {
	// 	return err
	// }
	var stateProof StateProof
	err := cdc.Unmarshal(proof, &stateProof)
	if err != nil {
		return sdkerrors.Wrap(err, "proof couldn't be decoded into StateProof struct")
	}

	consensusState, err := GetConsensusState(clientStore, cdc, height)
	if err != nil {
		return err
	}

	trie, err := trieproof.BuildTrie(stateProof.Proofs, consensusState.Root)
	if err != nil {
		logger.Sugar().Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyPacketReceiptAbsence()", "failure to build trie proof: ", err)

		return err
	}
	// try to find value
	value := trie.Get(stateProof.Key)
	if value != nil {
		logger.Sugar().Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyPacketReceiptAbsence()", "found value : ", value)

		return sdkerrors.Wrap(errors.New("VerifyPacketReceiptAbsence error "), "found value")
	}

	return nil
}

// verifyDelayPeriodPassed will ensure that at least delayTimePeriod amount of time and delayBlockPeriod number of blocks have passed
// since consensus state was submitted before allowing verification to continue.
func verifyDelayPeriodPassed(ctx sdk.Context, store sdk.KVStore, proofHeight exported.Height, delayTimePeriod, delayBlockPeriod uint64) error {
	// check that executing chain's timestamp has passed consensusState's processed time + delay time period
	/*
		processedTime, ok := GetProcessedTime(store, proofHeight)
		if !ok {
			return sdkerrors.Wrapf(ErrProcessedTimeNotFound, "processed time not found for height: %s", proofHeight)
		}
		currentTimestamp := uint64(ctx.BlockTime().UnixNano())
		validTime := processedTime + delayTimePeriod
		// NOTE: delay time period is inclusive, so if currentTimestamp is validTime, then we return no error
		if currentTimestamp < validTime {
			return sdkerrors.Wrapf(ErrDelayPeriodNotPassed, "cannot verify packet until time: %d, current time: %d",
				validTime, currentTimestamp)
		}
		// check that executing chain's height has passed consensusState's processed height + delay block period
		processedHeight, ok := GetProcessedHeight(store, proofHeight)
		if !ok {
			return sdkerrors.Wrapf(ErrProcessedHeightNotFound, "processed height not found for height: %s", proofHeight)
		}
		currentHeight := clienttypes.GetSelfHeight(ctx)
		validHeight := clienttypes.NewHeight(processedHeight.GetRevisionNumber(), processedHeight.GetRevisionHeight()+delayBlockPeriod)
		// NOTE: delay block period is inclusive, so if currentHeight is validHeight, then we return no error
		if currentHeight.LT(validHeight) {
			return sdkerrors.Wrapf(ErrDelayPeriodNotPassed, "cannot verify packet until height: %s, current height: %s",
				validHeight, currentHeight)
		}
	*/
	return nil
}
