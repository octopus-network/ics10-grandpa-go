package grandpa

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	// "github.com/tendermint/tendermint/types/time"
)

// TODO: VerifyUpgradeAndUpdateState
// VerifyUpgradeAndUpdateState checks if the upgraded client has been committed by the current client
// It will zero out all client-specific fields (e.g. TrustingPeriod and verify all data
// in client state that must be the same across all valid Tendermint clients for the new chain.
// VerifyUpgrade will return an error if:
// - the upgradedClient is not a Tendermint ClientState
// - the lastest height of the client state does not have the same revision number or has a greater
// height than the committed client.
// - the height of upgraded client is not greater than that of current client
// - the latest height of the new client does not match or is greater than the height in committed client
// - any Tendermint chain specified parameter in upgraded client such as ChainID, UnbondingPeriod,
//   and ProofSpecs do not match parameters set by committed client
// func (cs ClientState) VerifyUpgradeAndUpdateState(
// 	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
// 	upgradedClient exported.ClientState, upgradedConsState exported.ConsensusState,
// 	proofUpgradeClient, proofUpgradeConsState []byte,
// ) (exported.ClientState, exported.ConsensusState, error) {
// 	ctx.Logger().Debug("LightClient:", "10-Grandpa", "method:", "VerifyUpgradeAndUpdateState")

// 	// last height of current counterparty chain must be client's latest height
// 	lastHeight := cs.GetLatestHeight()

// 	if !upgradedClient.GetLatestHeight().GT(lastHeight) {
// 		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "upgraded client height %s must be at greater than current client height %s",
// 			upgradedClient.GetLatestHeight(), lastHeight)
// 	}

// 	// upgraded client state and consensus state must be IBC tendermint client state and consensus state
// 	// this may be modified in the future to upgrade to a new IBC tendermint type
// 	// counterparty must also commit to the upgraded consensus state at a sub-path under the upgrade path specified
// 	gpUpgradeClient, ok := upgradedClient.(*ClientState)
// 	if !ok {
// 		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "upgraded client must be Tendermint client. expected: %T got: %T",
// 			&ClientState{}, upgradedClient)
// 	}

// 	// unmarshal proofs
// 	var merkleProofClient, merkleProofConsState commitmenttypes.MerkleProof
// 	if err := cdc.Unmarshal(proofUpgradeClient, &merkleProofClient); err != nil {
// 		return nil, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal client merkle proof: %v", err)
// 	}
// 	if err := cdc.Unmarshal(proofUpgradeConsState, &merkleProofConsState); err != nil {
// 		return nil, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal consensus state merkle proof: %v", err)
// 	}

// 	// Construct new client state and consensus state
// 	// Relayer chosen client parameters are ignored.
// 	// All chain-chosen parameters come from committed client, all client-chosen parameters
// 	// come from current client.
// 	newClientState := NewClientState(
// 		gpUpgradeClient.ChainType, gpUpgradeClient.ChainId, gpUpgradeClient.ParachainId, gpUpgradeClient.BeefyActivationBlock,
// 		gpUpgradeClient.LatestBeefyHeight, gpUpgradeClient.MmrRootHash, gpUpgradeClient.LatestChainHeight,
// 		gpUpgradeClient.FrozenHeight, gpUpgradeClient.AuthoritySet, gpUpgradeClient.NextAuthoritySet,
// 	)

// 	if err := newClientState.Validate(); err != nil {
// 		return nil, nil, sdkerrors.Wrap(err, "updated client state failed basic validation")
// 	}

// 	// The new consensus state is merely used as a trusted kernel against which headers on the new
// 	// chain can be verified. The root is just a stand-in sentinel value as it cannot be known in advance, thus no proof verification will pass.
// 	// The timestamp and the NextValidatorsHash of the consensus state is the blocktime and NextValidatorsHash
// 	// of the last block committed by the old chain. This will allow the first block of the new chain to be verified against
// 	// the last validators of the old chain so long as it is submitted within the TrustingPeriod of this client.
// 	// NOTE: We do not set processed time for this consensus state since this consensus state should not be used for packet verification
// 	// as the root is empty. The next consensus state submitted using update will be usable for packet-verification.
// 	newConsState := NewConsensusState(

// 		[]byte{},
// 		time.Unix(0, 0),
// 	)

// 	// set metadata for this consensus state
// 	setConsensusMetadata(ctx, clientStore, clienttypes.NewHeight(0, uint64(gpUpgradeClient.LatestChainHeight)))

// 	return newClientState, newConsState, nil
// }

// CheckSubstituteAndUpdateState will try to update the client with the state of the
// substitute.
//
// AllowUpdateAfterMisbehaviour and AllowUpdateAfterExpiry have been deprecated.
// Please see ADR 026 for more information.
//
// The following must always be true:
//   - The substitute client is the same type as the subject client
//   - The subject and substitute client states match in all parameters (expect frozen height, latest height, and chain-id)
//
// In case 1) before updating the client, the client will be unfrozen by resetting
// the FrozenHeight to the zero Height.
func (cs ClientState) VerifyUpgradeAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	upgradedClient exported.ClientState, upgradedConsState exported.ConsensusState,
	proofUpgradeClient, proofUpgradeConsState []byte,
) error {
	panic("")
}
