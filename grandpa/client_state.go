package grandpa

import (
	"errors"
	"time"

	ics23 "github.com/confio/ics23/go"
	"github.com/octopus-network/beefy-go/beefy"

	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	trieproof "github.com/octopus-network/trie-go/trie/proof"
)

var _ exported.ClientState = (*ClientState)(nil)

// NewClientState creates a new ClientState instance
func NewClientState(
	chainType uint32,
	chainId string,
	parachainId uint32,
	beefyActivationBlock uint32,
	latestBeefyHeight uint32,
	mmrRootHash []byte,
	latestHeight uint32,
	frozenHeight uint32,
	authoritySet BeefyAuthoritySet,
	nextAuthoritySet BeefyAuthoritySet,
) *ClientState {
	return &ClientState{
		ChainType:            chainType,
		ChainId:              chainId,
		ParachainId:          parachainId,
		BeefyActivationBlock: beefyActivationBlock,
		LatestBeefyHeight:    latestBeefyHeight,
		MmrRootHash:          mmrRootHash,
		LatestChainHeight:    latestHeight,
		FrozenHeight:         frozenHeight,
		AuthoritySet:         authoritySet,
		NextAuthoritySet:     nextAuthoritySet,
	}
}

// GetChainID returns the chain-id
func (cs ClientState) GetChainID() string {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.GetChainID()")

	return cs.ChainId
}

// ClientType is tendermint.
func (cs ClientState) ClientType() string {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.ClientType()")

	return ModuleName
}

// TODO: which height? latest beefy height or latest chain height.
func (cs ClientState) GetLatestHeight() exported.Height {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.GetLatestHeight()")

	return clienttypes.Height{
		RevisionNumber: 0,
		// RevisionHeight: uint64(cs.LatestBeefyHeight),
		RevisionHeight: uint64(cs.LatestChainHeight),
	}
}

// GetTimestampAtHeight returns the timestamp in nanoseconds of the consensus state at the given height.
func (cs ClientState) GetTimestampAtHeight(
	ctx sdk.Context,
	clientStore sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
) (uint64, error) {
	// get consensus state at height from clientStore to check for expiry
	// consState, found := GetConsensusState(clientStore, cdc, height)
	// if !found {
	// 	return 0, sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "height (%s)", height)
	// }
	// return consState.GetTimestamp(), nil
	panic("")
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
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.Status()")

	if cs.FrozenHeight > 0 {
		return exported.Frozen
	}

	return exported.Active
}

// TODO: check expired
// IsExpired returns whether or not the client has passed the trusting period since the last
// update (in which case no headers are considered valid).
func (cs ClientState) IsExpired(latestTimestamp, now time.Time) bool {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.IsExpired()")

	return false
}

// Validate performs a basic validation of the client state fields.
func (cs ClientState) Validate() error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.Validate()")

	if cs.LatestBeefyHeight == 0 {

		return sdkerrors.Wrap(ErrInvalidHeaderHeight, "beefy height cannot be zero")
	}

	return nil
}

// GetProofSpecs returns the format the client expects for proof verification
// as a string array specifying the proof type for each position in chained proof
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.GetProofSpecs()")

	ps := []*ics23.ProofSpec{}
	// ps := commitmenttypes.GetSDKSpecs()
	return ps

}

// ZeroCustomFields returns a ClientState that is a copy of the current ClientState
// with all client customizable fields zeroed out
func (cs ClientState) ZeroCustomFields() exported.ClientState {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.ZeroCustomFields()")

	// copy over all chain-specified fields
	// and leave custom fields empty

	return &ClientState{
		ChainId:              cs.ChainId,
		ChainType:            cs.ChainType,
		BeefyActivationBlock: cs.BeefyActivationBlock,
		LatestBeefyHeight:    cs.LatestBeefyHeight,
		MmrRootHash:          cs.MmrRootHash,
		LatestChainHeight:    cs.LatestChainHeight,
		FrozenHeight:         cs.FrozenHeight,
		AuthoritySet:         cs.AuthoritySet,
		NextAuthoritySet:     cs.NextAuthoritySet,
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
	//Note,this height must be solochain or parachain height
	latestChainHeigh := clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: uint64(cs.LatestChainHeight),
	}
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
	panic("")
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
	panic("")
}

// VerifyClientState verifies a proof of the client state of the running chain
// stored on the target machine
func (cs ClientState) VerifyClientState(
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	prefix exported.Prefix,
	counterpartyClientIdentifier string,
	proof []byte,
	clientState exported.ClientState,
) error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyClientState()",
		"height", height, "prefix", prefix, "counterpartyClientIdentifier", "proof", "clientState", clientState)

	if clientState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidClient, "client state cannot be empty")
	}

	// asset tendermint clientstate
	_, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "invalid client type %T, expected %T", clientState, &ClientState{})
	}

	// clientPrefixedPath := commitmenttypes.NewMerklePath(host.FullClientStatePath(counterpartyClientIdentifier))
	// path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	// if err != nil {
	// 	return err
	// }

	// build state key,I think don`t build state key for substrate state proof
	// just upload in proof
	// key:=beefy.CreateStorageKeyPrefix(prefix, method string)
	// the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	encodedClientState, err := gsrpccodec.Encode(clientState)
	if err != nil {
		return sdkerrors.Wrap(err, "clientState could not be scale encoded")
	}
	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyClientState()", "encodedClientState", encodedClientState)
	// get state proof
	stateProof, provingConsensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// err = beefy.VerifyStateProof(stateProof.Proofs, provingConsensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, provingConsensusState.Root, stateProof.Key, encodedClientState)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyClientState()", "failure to verify state proof: ", err)
		return err
	}

	return nil

}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Tendermint client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	counterpartyClientIdentifier string,
	consensusHeight exported.Height,
	prefix exported.Prefix,
	proof []byte,
	consensusState exported.ConsensusState,
) error {
	// Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyClientConsensusState()")
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyClientConsensusState()",
		"height", height, "consensusHeight", consensusHeight, "prefix", prefix, "counterpartyClientIdentifier", "proof", "consensusState", consensusState)
	if consensusState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidClient, "consensus state cannot be empty")
	}

	// asset tendermint consensuse state
	_, ok := consensusState.(*ibctmtypes.ConsensusState)
	if !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "invalid client type %T, expected %T", consensusState, &ConsensusState{})
	}

	// encode consensue state
	encodedConsensueState, err := gsrpccodec.Encode(consensusState)
	if err != nil {
		return sdkerrors.Wrap(err, "consensusState could not be scale encoded")
	}
	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyClientConsensusState()", "encodedConsensueState", encodedConsensueState)

	// clientPrefixedPath := commitmenttypes.NewMerklePath(host.FullConsensusStatePath(counterpartyClientIdentifier, consensusHeight))
	// path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	// if err != nil {
	// 	return err
	// }

	// the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	stateProof, provingConsensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// verify state proof
	// err = beefy.VerifyStateProof(stateProof.Proofs, provingConsensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, provingConsensusState.Root, stateProof.Key, encodedConsensueState)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyClientConsensusState()", "failure to verify state proof: ", err)
		return err
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	connectionID string,
	connectionEnd exported.ConnectionI,
) error {
	// Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyConnectionState()")

	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyConnectionState()",
		"height", height, "prefix", prefix, "proof", proof, "connectionID", connectionID, "connectionEnd", connectionEnd)

	stateProof, consensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// connectionPath := commitmenttypes.NewMerklePath(host.ConnectionPath(connectionID))
	// path, err := commitmenttypes.ApplyPrefix(prefix, connectionPath)
	// if err != nil {
	// 	return err
	// }
	// the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	connection, ok := connectionEnd.(connectiontypes.ConnectionEnd)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid connection type %T", connectionEnd)
	}

	//encode connectionend
	encodedConnEnd, err := gsrpccodec.Encode(connection)
	if err != nil {
		return sdkerrors.Wrap(err, "connection state could not be scale encoded")
	}
	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyConnectionState()", "encodedConnEnd", encodedConnEnd)

	// err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, encodedConnEnd)

	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyConnectionState()", "failure to verify state proof: ", err)
		return err
	}

	return nil

}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	channel exported.ChannelI,
) error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyChannelState()",
		"height", height, "prefix", prefix, "proof", proof, "portID", portID, "channelID", channelID, "channel", channel)

	stateProof, consensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// channelPath := commitmenttypes.NewMerklePath(host.ChannelPath(portID, channelID))
	// path, err := commitmenttypes.ApplyPrefix(prefix, channelPath)
	// if err != nil {
	// 	return err
	// }

	// the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	channelEnd, ok := channel.(channeltypes.Channel)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid channel type %T", channel)
	}

	//encode channel end
	encodedChanEnd, err := gsrpccodec.Encode(channelEnd)
	if err != nil {
		return sdkerrors.Wrap(err, "channel end could not be scale encoded")
	}

	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyChannelState()", "encodedChanEnd", encodedChanEnd)

	// err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, encodedChanEnd)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyChannelState()", "failure to verify state proof: ", err)

		return err
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	ctx sdk.Context,
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	delayTimePeriod uint64,
	delayBlockPeriod uint64,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
) error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyPacketCommitment()",
		"height", height, "prefix", prefix, "proof", proof, "portID", portID, "channelID", channelID,
		"delayTimePeriod", delayTimePeriod, "delayBlockPeriod", delayBlockPeriod, "sequence", sequence,
		"commitmentBytes", commitmentBytes)

	stateProof, consensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// check delay period has passed
	if err := verifyDelayPeriodPassed(ctx, store, height, delayTimePeriod, delayBlockPeriod); err != nil {
		return err
	}

	// check delay period has passed
	// if err := verifyDelayPeriodPassed(ctx, store, height, delayTimePeriod, delayBlockPeriod); err != nil {
	// 	return err
	// }

	// commitmentPath := commitmenttypes.NewMerklePath(host.PacketCommitmentPath(portID, channelID, sequence))
	// path, err := commitmenttypes.ApplyPrefix(prefix, commitmentPath)
	// if err != nil {
	// 	return err
	// }

	// the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	// TODO:confirm the commitmentBytes is scale encode?
	encodedCZ, err := gsrpccodec.Encode(commitmentBytes)
	if err != nil {
		return sdkerrors.Wrap(err, "commitmentBytes could not be scale encoded")
	}

	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyPacketCommitment()", "encoded commitmentBytes", encodedCZ)

	// err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, encodedCZ)

	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyPacketCommitment()", "failure to verify state proof: ", err)
		return err
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	delayTimePeriod uint64,
	delayBlockPeriod uint64,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyPacketAcknowledgement()",
		"height", height, "prefix", prefix, "proof", proof, "portID", portID, "channelID", channelID,
		"delayTimePeriod", delayTimePeriod, "delayBlockPeriod", delayBlockPeriod, "sequence", sequence,
		"acknowledgement", acknowledgement)

	stateProof, consensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// check delay period has passed
	if err := verifyDelayPeriodPassed(ctx, store, height, delayTimePeriod, delayBlockPeriod); err != nil {
		return err
	}

	// ackPath := commitmenttypes.NewMerklePath(host.PacketAcknowledgementPath(portID, channelID, sequence))
	// path, err := commitmenttypes.ApplyPrefix(prefix, ackPath)
	// if err != nil {
	// 	return err
	// }

	// // the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	// acknowledgement
	// TODO: confirm the acknowledgement is scale encode?
	encodedAck, err := gsrpccodec.Encode(acknowledgement)
	if err != nil {
		return sdkerrors.Wrap(err, "acknowledgement could not be scale encoded")
	}

	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyPacketAcknowledgement()", "encoded acknowledgement", encodedAck)

	// err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, encodedAck)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyPacketAcknowledgement()", "failure to verify state proof: ", err)
		return err
	}
	return nil
}

// TODO: confirm process
// VerifyPacketReceiptAbsence verifies a proof of the absence of an
// incoming packet receipt at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketReceiptAbsence(
	ctx sdk.Context,
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	delayTimePeriod uint64,
	delayBlockPeriod uint64,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
) error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyPacketReceiptAbsence()",
		"height", height, "prefix", prefix, "proof", proof, "portID", portID, "channelID", channelID,
		"delayTimePeriod", delayTimePeriod, "delayBlockPeriod", delayBlockPeriod, "sequence", sequence)

	stateProof, consensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}
	// check delay period has passed
	if err := verifyDelayPeriodPassed(ctx, store, height, delayTimePeriod, delayBlockPeriod); err != nil {
		return err
	}

	// receiptPath := commitmenttypes.NewMerklePath(host.PacketReceiptPath(portID, channelID, sequence))
	// _, err = commitmenttypes.ApplyPrefix(prefix, receiptPath)
	// if err != nil {
	// 	return err
	// }
	// path, err := commitmenttypes.ApplyPrefix(prefix, receiptPath)
	// if err != nil {
	// 	return err
	// }

	// // the key is just the raw utf-8 bytes of the prefix + path
	// key := []byte(strings.Join(path.GetKeyPath(), ""))
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "keyPath could not be scale encoded")
	// }

	// TODO: asset stateProof.Key == beefy.CreateStorageKeyPrefix(path,method)

	// build trie proof tree
	trie, err := trieproof.BuildTrie(stateProof.Proofs, consensusState.Root)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyPacketReceiptAbsence()", "failure to build trie proof: ", err)

		return err
	}
	// find value
	value := trie.Get(stateProof.Key)
	if value != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyPacketReceiptAbsence()", "found value : ", value)

		return sdkerrors.Wrap(errors.New("VerifyPacketReceiptAbsence error "), "found value")
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	ctx sdk.Context,
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	height exported.Height,
	delayTimePeriod uint64,
	delayBlockPeriod uint64,
	prefix exported.Prefix,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	Logger.Debug("LightClient:", "10-Grandpa", "method:", "ClientState.VerifyNextSequenceRecv()",
		"height", height, "prefix", prefix, "proof", proof, "portID", portID, "channelID", channelID,
		"delayTimePeriod", delayTimePeriod, "delayBlockPeriod", delayBlockPeriod, "	nextSequenceRecv", nextSequenceRecv)

	stateProof, consensusState, err := produceVerificationArgs(store, cdc, cs, height, prefix, proof)
	if err != nil {
		return err
	}

	// check delay period has passed
	if err := verifyDelayPeriodPassed(ctx, store, height, delayTimePeriod, delayBlockPeriod); err != nil {
		return err
	}
	// nextSequenceRecvPath := commitmenttypes.NewMerklePath(host.NextSequenceRecvPath(portID, channelID))
	// path, err := commitmenttypes.ApplyPrefix(prefix, nextSequenceRecvPath)
	// if err != nil {
	// 	return err
	// }

	// key, err := scale.Encode(path)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "next sequence recv path could not be scale encoded")
	// }
	// TODO: confirm the nextSequenceRecv is scale encode or Uint64ToBigEndian?
	// bz := sdk.Uint64ToBigEndian(nextSequenceRecv)
	//encode channel end
	encodedNextSequenceRecv, err := gsrpccodec.Encode(nextSequenceRecv)
	if err != nil {
		return sdkerrors.Wrap(err, "channel end could not be scale encoded")
	}
	Logger.Debug("LightClient:", "10-Grandpa", "method:",
		"ClientState.VerifyNextSequenceRecv()", "sdk.Uint64ToBigEndian(nextSequenceRecv)", encodedNextSequenceRecv)

	// err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, encodedNextSequenceRecv)

	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:",
			"ClientState.VerifyNextSequenceRecv()", "failure to verify state proof: ", err)

		return err
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

// produceVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions and returns the unmarshalled
// merkle proof, the consensus state and an error if one occurred.
func produceVerificationArgs(
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	cs ClientState,
	height exported.Height,
	prefix exported.Prefix,
	proof []byte,
) (stateProof StateProof, consensusState *ConsensusState, err error) {
	// if cs.GetLatestHeight().LT(height) {
	// 	return commitmenttypes.MerkleProof{}, nil, sdkerrors.Wrapf(
	// 		sdkerrors.ErrInvalidHeight,
	// 		"client state height < proof height (%d < %d), please ensure the client has been updated", cs.GetLatestHeight(), height,
	// 	)
	// }

	// no height checks because parachain_header height fits into revision_number
	// so if fetching consensus state fails, we don't have the consensus state for the parachain header.

	if proof == nil {
		return StateProof{}, nil, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof cannot be empty")
	}

	if prefix == nil {
		return StateProof{}, nil, sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(*commitmenttypes.MerklePrefix)
	if !ok {
		return StateProof{}, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected *MerklePrefix", prefix)
	}

	// Note: decode proof
	// err = gsrpccodec.Decode(proof, &stateProof)
	err = cdc.Unmarshal(proof, &stateProof)
	if err != nil {
		return StateProof{}, nil, sdkerrors.Wrap(err, "proof couldn't be decoded into StateProof struct")
	}

	consensusState, err = GetConsensusState(store, cdc, height)
	if err != nil {
		return StateProof{}, nil, sdkerrors.Wrap(err, "please ensure the proof was constructed against a height that exists on the client")
	}

	return stateProof, consensusState, nil
}
