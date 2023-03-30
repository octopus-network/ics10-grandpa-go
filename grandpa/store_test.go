package grandpa_test

import (
	"math"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (suite *GrandpaTestSuite) TestGetConsensusState() {
	suite.Suite.T().Skip("TestGetConsensusState")
}

func (suite *GrandpaTestSuite) TestGetProcessedTime() {
	suite.Suite.T().Skip("TestGetProcessedTime")
}

func (suite *GrandpaTestSuite) TestIterationKey() {
	testHeights := []exported.Height{
		clienttypes.NewHeight(0, 1),
		clienttypes.NewHeight(0, 1234),
		clienttypes.NewHeight(7890, 4321),
		clienttypes.NewHeight(math.MaxUint64, math.MaxUint64),
	}
	for _, h := range testHeights {
		k := tendermint.IterationKey(h)
		retrievedHeight := tendermint.GetHeightFromIterationKey(k)
		suite.Require().Equal(h, retrievedHeight, "retrieving height from iteration key failed")
	}
}

func (suite *GrandpaTestSuite) TestIterateConsensusStates() {
	suite.Suite.T().Skip("TestIterateConsensusStates")
}

func (suite *GrandpaTestSuite) TestGetNeighboringConsensusStates() {
	suite.Suite.T().Skip("TestGetNeighboringConsensusStates")
}
