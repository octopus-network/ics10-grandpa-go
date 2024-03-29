package grandpa

import (
	"reflect"
	time "time"

	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/octopus-network/beefy-go/beefy"
)

var _ exported.ClientMessage = &Header{}

// ConsensusState returns the updated consensus state associated with the header
func (h Header) ConsensusState() *ConsensusState {
	// get latest header and time
	_, latestHeader, latestTime, err := getLastestChainHeader(&h)
	if err != nil {
		logger.Sugar().Error("LightClient:", "10-Grandpa", "method:", "getLastestHeader error: ", err)
		return nil
	}
	// build consensue state
	return &ConsensusState{
		Root:      latestHeader.StateRoot[:],
		Timestamp: latestTime,
	}
}

// ClientType defines that the Header is a grandpa consensus algorithm
func (h Header) ClientType() string {
	return ModuleName
}

// GetHeight returns the current height. It returns 0 if the grandpa
// header is nil.
// NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetHeight() exported.Height {
	// use the beefy height as header height
	// latestBeefyHeight := h.BeefyMmr.SignedCommitment.Commitment.BlockNumber
	// get latest header and time
	chainID, latestHeader, _, err := getLastestChainHeader(&h)
	if err != nil {
		logger.Sugar().Error("LightClient:", "10-Grandpa", "method:", "GetHeight error: ", err)
		return nil
	}

	revision := clienttypes.ParseChainID(chainID)
	return clienttypes.NewHeight(revision, uint64(latestHeader.Number))
}

// // GetTime returns the current block timestamp. It returns a zero time if
// // the tendermint header is nil.
// // NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetTime() time.Time {
	// get latest header and time
	_, _, latestTime, err := getLastestChainHeader(&h)
	logger.Sugar().Debug("latestTime:", latestTime)
	if err != nil {
		logger.Sugar().Error("LightClient:", "10-Grandpa", "method:", "getLastestHeader error: ", err)
		return time.UnixMilli(0)
	}
	return latestTime
}

func getLastestChainHeader(h *Header) (string, gsrpctypes.Header, time.Time, error) {
	var chainID string
	var latestHeader gsrpctypes.Header
	var latestTimestamp time.Time
	var latestHeight uint32
	headerMessage := h.GetMessage()

	switch headerMap := headerMessage.(type) {
	case *Header_SubchainHeaderMap:
		subchainHeaderMap := headerMap.SubchainHeaderMap.SubchainHeaderMap
		for num := range subchainHeaderMap {
			if latestHeight < num {
				latestHeight = num
			}
		}
		subchainHeader := subchainHeaderMap[latestHeight]
		chainID = subchainHeader.ChainId

		// var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(subchainHeader.BlockHeader, &latestHeader)
		if err != nil {
			return "", latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}
		logger.Sugar().Debug("decodeHeader.Number:", latestHeader.Number)

		// verify timestamp and get it
		err = beefy.VerifyStateProof(subchainHeader.Timestamp.Proofs,
			latestHeader.StateRoot[:], subchainHeader.Timestamp.Key,
			subchainHeader.Timestamp.Value)
		if err != nil {
			logger.Sugar().Error("LightClient:", "10-Grandpa", "method:", "VerifyStateProof error: ", err)
			return "", latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		// decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(subchainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			logger.Sugar().Error("decode timestamp error:", err)
			return "", latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	case *Header_ParachainHeaderMap:
		parachainHeaderMap := headerMap.ParachainHeaderMap.ParachainHeaderMap
		for num := range parachainHeaderMap {
			if latestHeight < num {
				latestHeight = num
			}
		}
		parachainHeader := parachainHeaderMap[latestHeight]
		chainID = parachainHeader.ChainId
		// var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(parachainHeader.BlockHeader, &latestHeader)
		if err != nil {
			return "", latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}

		// verify timestamp and get it
		err = beefy.VerifyStateProof(parachainHeader.Timestamp.Proofs,
			latestHeader.StateRoot[:], parachainHeader.Timestamp.Key,
			parachainHeader.Timestamp.Value)
		if err != nil {
			logger.Sugar().Error("LightClient:", "10-Grandpa", "method:", "VerifyStateProof error: ", err)
			return "", latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		// decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(parachainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			logger.Sugar().Error("decode timestamp error:", err)
			return "", latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	}

	return chainID, latestHeader, latestTimestamp, nil
}

// ValidateBasic calls the header ValidateBasic function and checks
// with MsgCreateClient
func (h Header) ValidateBasic() error {
	if reflect.DeepEqual(h, Header{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Grandpa header cannot be nil")
	}

	if reflect.DeepEqual(h.BeefyMmr, BeefyMMR{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "beefy mmr cannot be nil")
	}

	if h.Message == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Header Message cannot be nil")
	}

	return nil
}
