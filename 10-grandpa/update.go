package grandpa

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

// VerifyClientMessage checks if the clientMessage is of type Header or Misbehaviour and verifies the message
func (cs *ClientState) VerifyClientMessage(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	clientMsg exported.ClientMessage,
) error {
	switch msg := clientMsg.(type) {
	case *Header:
		return cs.verifyHeader(ctx, clientStore, cdc, msg)
	default:
		return clienttypes.ErrInvalidClientType
	}
}

// verifyHeader returns an error if:
// - the client or header provided are not parseable
// - the header is invalid
// - header height is less than or equal to the trusted header height
// - header revision is not equal to trusted header revision
// - header valset commit verification fails
// - header timestamp is past the trusting period in relation to the consensus state
// - header timestamp is less than or equal to the consensus state timestamp
func (cs *ClientState) verifyHeader(
	ctx sdk.Context, clientStore sdk.KVStore, cdc codec.BinaryCodec,
	header *Header,
) error {

	return nil
}

func (cs ClientState) UpdateState(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, clientMsg exported.ClientMessage) []exported.Height {
	header, _ := clientMsg.(*Header)
	height := header.GetHeight().(clienttypes.Height)
	return []exported.Height{height}
}

// pruneOldestConsensusState will retrieve the earliest consensus state for this clientID and check if it is expired. If it is,
// that consensus state will be pruned from store along with all associated metadata. This will prevent the client store from
// becoming bloated with expired consensus states that can no longer be used for updates and packet verification.
func (cs ClientState) pruneOldestConsensusState(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore) {
	// Check the earliest consensus state to see if it is expired, if so then set the prune height
	// so that we can delete consensus state and all associated metadata.
	// var (
	// 	pruneHeight exported.Height
	// )

	// pruneCb := func(height exported.Height) bool {
	// 	consState, found := GetConsensusState(clientStore, cdc, height)
	// 	// this error should never occur
	// 	if !found {
	// 		panic(sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "failed to retrieve consensus state at height: %s", height))
	// 	}

	// 	if cs.IsExpired(consState.Timestamp, ctx.BlockTime()) {
	// 		pruneHeight = height
	// 	}

	// 	return true
	// }

	// IterateConsensusStateAscending(clientStore, pruneCb)

	// // if pruneHeight is set, delete consensus state and metadata
	// if pruneHeight != nil {
	// 	deleteConsensusState(clientStore, pruneHeight)
	// 	deleteConsensusMetadata(clientStore, pruneHeight)
	// }
	panic("implement me")
}

// UpdateStateOnMisbehaviour updates state upon misbehaviour, freezing the ClientState. This method should only be called when misbehaviour is detected
// as it does not perform any misbehaviour checks.
func (cs ClientState) UpdateStateOnMisbehaviour(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, _ exported.ClientMessage) {
	clientStore.Set(host.ClientStateKey(), clienttypes.MustMarshalClientState(cdc, &cs))

	panic("implement me")
}
