package grandpa

import (
	time "time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// SentinelRoot is used as a stand-in root value for the consensus state set at the upgrade height
const SentinelRoot = "sentinel_root"

// NewConsensusState creates a new ConsensusState instance.
func NewConsensusState(
	root []byte,
	timestamp time.Time,
) *ConsensusState {
	return &ConsensusState{
		Root:      root,
		Timestamp: timestamp,
	}
}

// ClientType returns Grandpa
func (ConsensusState) ClientType() string {
	return ModuleName
}

// GetRoot returns the commitment Root for the specific
func (cs ConsensusState) GetRoot() exported.Root {
	return commitmenttypes.NewMerkleRoot([]byte(cs.Root))

}

// GetTimestamp returns block time in nanoseconds of the header that created consensus state
func (cs ConsensusState) GetTimestamp() uint64 {

	return uint64(cs.Timestamp.UnixMilli())

}

// ValidateBasic defines a basic validation for the tendermint consensus state.
// NOTE: ProcessedTimestamp may be zero if this is an initial consensus state passed in by relayer
// as opposed to a consensus state constructed by the chain.
func (cs ConsensusState) ValidateBasic() error {
	// if cs.Root.Empty() {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "root cannot be empty")
	// }
	if len(cs.Root) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "root cannot be empty")
	}

	if cs.Timestamp.UnixMilli() <= 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp must be a positive Unix time")
	}
	return nil
}
