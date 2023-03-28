package grandpa

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// CheckForMisbehaviour detects duplicate height misbehaviour and BFT time violation misbehaviour
// in a submitted Header message and verifies the correctness of a submitted Misbehaviour ClientMessage
func (cs ClientState) CheckForMisbehaviour(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, msg exported.ClientMessage) bool {
	panic("")
}

// CheckSubstituteAndUpdateState will try to update the client with the state of the
// substitute if and only if the proposal passes and one of the following conditions are
// satisfied:
//  1. AllowUpdateAfterMisbehaviour and Status() == Frozen
//  2. AllowUpdateAfterExpiry=true and Status() == Expired
//
// The following must always be true:
//   - The substitute client is the same type as the subject client
//   - The subject and substitute client states match in all parameters (expect frozen height, latest height, and chain-id)
//
// In case 1) before updating the client, the client will be unfrozen by resetting
// the FrozenHeight to the zero Height. If a client is frozen and AllowUpdateAfterMisbehaviour
// is set to true, the client will be unexpired even if AllowUpdateAfterExpiry is set to false.
// func (cs ClientState) CheckSubstituteAndUpdateState(
// 	ctx sdk.Context, cdc codec.BinaryCodec, subjectClientStore,
// 	substituteClientStore sdk.KVStore, substituteClient exported.ClientState,
// ) (exported.ClientState, error) {
// 	ctx.Logger().Debug("LightClient:", "10-Grandpa", "method:", "CheckSubstituteAndUpdateState")

// 	substituteClientState, ok := substituteClient.(*ClientState)
// 	if !ok {
// 		return nil, sdkerrors.Wrapf(
// 			clienttypes.ErrInvalidClient, "expected type %T, got %T", &ClientState{}, substituteClient,
// 		)
// 	}

// 	if !IsMatchingClientState(cs, *substituteClientState) {
// 		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidSubstitute, "subject client state does not match substitute client state")
// 	}

// 	if cs.Status(ctx, subjectClientStore, cdc) == exported.Frozen {
// 		// unfreeze the client
// 		cs.FrozenHeight = uint32(clienttypes.ZeroHeight().RevisionNumber)
// 	}

// 	// copy consensus states and processed time from substitute to subject
// 	// starting from initial height and ending on the latest height (inclusive)
// 	height := substituteClientState.GetLatestHeight()

// 	consensusState, err := GetConsensusState(substituteClientStore, cdc, height)
// 	if err != nil {
// 		return nil, sdkerrors.Wrap(err, "unable to retrieve latest consensus state for substitute client")
// 	}

// 	SetConsensusState(subjectClientStore, cdc, consensusState, height)

// 	// set metadata stored for the substitute consensus state
// 	processedHeight, found := GetProcessedHeight(substituteClientStore, height)
// 	if !found {
// 		return nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "unable to retrieve processed height for substitute client latest height")
// 	}

// 	processedTime, found := GetProcessedTime(substituteClientStore, height)
// 	if !found {
// 		return nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "unable to retrieve processed time for substitute client latest height")
// 	}

// 	setConsensusMetadataWithValues(subjectClientStore, height, processedHeight, processedTime)

// 	cs.LatestChainHeight = substituteClientState.LatestChainHeight
// 	cs.ChainId = substituteClientState.ChainId

// 	// no validation is necessary since the substitute is verified to be Active
// 	// in 02-client.

// 	return &cs, nil
// }

// // IsMatchingClientState returns true if all the client state parameters match
// // except for frozen height, latest height, and chain-id.
// func IsMatchingClientState(subject, substitute ClientState) bool {
// 	// zero out parameters which do not need to match
// 	subject.LatestChainHeight = uint32(clienttypes.ZeroHeight().RevisionHeight)
// 	subject.FrozenHeight = uint32(clienttypes.ZeroHeight().RevisionHeight)
// 	substitute.LatestChainHeight = uint32(clienttypes.ZeroHeight().RevisionHeight)
// 	substitute.FrozenHeight = uint32(clienttypes.ZeroHeight().RevisionHeight)
// 	subject.ChainId = ""
// 	substitute.ChainId = ""

// 	return reflect.DeepEqual(subject, substitute)
// }

func (cs ClientState) CheckSubstituteAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryCodec, subjectClientStore,
	substituteClientStore sdk.KVStore, substituteClient exported.ClientState,
) error {

	panic("")
}
