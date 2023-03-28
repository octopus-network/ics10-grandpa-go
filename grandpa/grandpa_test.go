package grandpa_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/octopus-network/beefy-go/beefy"
	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibcgptypes "github.com/octopus-network/ics10-grandpa-go/grandpa"
)

const (
	chainID                        = "sub"
	chainIDRevision0               = "sub-revision-0"
	chainIDRevision1               = "sub-revision-1"
	clientID                       = "submainnet"
	trustingPeriod   time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod        time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift    time.Duration = time.Second * 10
)

var (
	height          = clienttypes.NewHeight(0, 4)
	newClientHeight = clienttypes.NewHeight(1, 1)
	upgradePath     = []string{"upgrade", "upgradedIBCState"}
)

type GrandpaTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	// TODO: deprecate usage in favor of testing package
	ctx          sdk.Context
	cdc          codec.Codec
	privVal      tmtypes.PrivValidator
	valSet       *tmtypes.ValidatorSet
	signers      map[string]tmtypes.PrivValidator
	valsHash     tmbytes.HexBytes
	header       *ibctmtypes.Header
	now          time.Time
	headerTime   time.Time
	clientTime   time.Time
	clientStates []*ibcgptypes.ClientState
}

func (suite *GrandpaTestSuite) initClientStates() {
	cs1 := ibcgptypes.ClientState{
		ChainId:              "sub-0",
		ChainType:            beefy.CHAINTYPE_SOLOCHAIN,
		BeefyActivationBlock: beefy.BEEFY_ACTIVATION_BLOCK,
		LatestBeefyHeight:    108,
		MmrRootHash:          []uint8{0x7b, 0xd6, 0x16, 0x8a, 0x24, 0xd1, 0xc1, 0xcf, 0x65, 0xca, 0x6e, 0x6e, 0xfe, 0x71, 0xab, 0x16, 0x95, 0xf9, 0x90, 0xaf, 0x1d, 0xaa, 0x11, 0x92, 0x4, 0xfa, 0xe9, 0x2d, 0xfb, 0x96, 0xfd, 0xa4},
	}
	suite.clientStates = append(suite.clientStates, &cs1)

}

func (suite *GrandpaTestSuite) SetupTest() {
	// suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	// suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	// suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	// // commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	// suite.coordinator.CommitNBlocks(suite.chainA, 2)
	// suite.coordinator.CommitNBlocks(suite.chainB, 2)

	// // TODO: deprecate usage in favor of testing package
	// checkTx := false
	// app := simapp.Setup(checkTx)

	// suite.cdc = app.AppCodec()

	// // now is the time of the current chain, must be after the updating header
	// // mocks ctx.BlockTime()
	// suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	// suite.clientTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// // Header time is intended to be time for any new header used for updates
	// suite.headerTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	// suite.privVal = ibctestingmock.NewPV()

	// pubKey, err := suite.privVal.GetPubKey()
	// suite.Require().NoError(err)

	// heightMinus1 := clienttypes.NewHeight(0, height.RevisionHeight-1)

	// val := tmtypes.NewValidator(pubKey, 10)
	// suite.signers = make(map[string]tmtypes.PrivValidator)
	// suite.signers[val.Address.String()] = suite.privVal
	// suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	// suite.valsHash = suite.valSet.Hash()
	// // //TODO: change to grandpa
	// suite.header = suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight), heightMinus1, suite.now, suite.valSet, suite.valSet, suite.valSet, suite.signers)
	// suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, Time: suite.now})
}

func getAltSigners(altVal *tmtypes.Validator, altPrivVal tmtypes.PrivValidator) map[string]tmtypes.PrivValidator {
	return map[string]tmtypes.PrivValidator{altVal.Address.String(): altPrivVal}
}

func getBothSigners(suite *GrandpaTestSuite, altVal *tmtypes.Validator, altPrivVal tmtypes.PrivValidator) (*tmtypes.ValidatorSet, map[string]tmtypes.PrivValidator) {
	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := map[string]tmtypes.PrivValidator{
		suiteVal.Address.String(): suite.privVal,
		altVal.Address.String():   altPrivVal,
	}
	return bothValSet, bothSigners
}

func TestGrandpaTestSuite(t *testing.T) {
	suite.Run(t, new(GrandpaTestSuite))
}
