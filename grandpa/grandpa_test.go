package grandpa_test

import (
	"testing"
	"time"

	"github.com/octopus-network/beefy-go/beefy"
	"github.com/stretchr/testify/suite"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibcgptypes "github.com/octopus-network/ics10-grandpa-go/grandpa"
)

const (
	chainID                        = "sub-0"
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

	coordinator  *ibctesting.Coordinator
	clientStates []*ibcgptypes.ClientState
}

func (suite *GrandpaTestSuite) initClientStates() {
	cs1 := ibcgptypes.ClientState{
		ChainId:               chainID,
		ChainType:             beefy.CHAINTYPE_SOLOCHAIN,
		BeefyActivationHeight: beefy.BEEFY_ACTIVATION_BLOCK,
		LatestBeefyHeight:     clienttypes.NewHeight(clienttypes.ParseChainID(chainID), 108),
		MmrRootHash:           []uint8{0x7b, 0xd6, 0x16, 0x8a, 0x24, 0xd1, 0xc1, 0xcf, 0x65, 0xca, 0x6e, 0x6e, 0xfe, 0x71, 0xab, 0x16, 0x95, 0xf9, 0x90, 0xaf, 0x1d, 0xaa, 0x11, 0x92, 0x4, 0xfa, 0xe9, 0x2d, 0xfb, 0x96, 0xfd, 0xa4},
	}
	suite.clientStates = append(suite.clientStates, &cs1)
}

func (suite *GrandpaTestSuite) SetupTest() {
	// suite.Suite.T().Skip("SetupTest")
}

func TestGrandpaTestSuite(t *testing.T) {
	suite.Run(t, new(GrandpaTestSuite))
}
