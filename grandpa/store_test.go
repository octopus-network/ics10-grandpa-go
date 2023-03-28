package grandpa_test

import (
	"math"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	types "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

func (suite *GrandpaTestSuite) TestGetConsensusState() {
	// var (
	// 	height exported.Height
	// 	path   *ibctesting.Path
	// )

	// testCases := []struct {
	// 	name     string
	// 	malleate func()
	// 	expPass  bool
	// }{
	// 	{
	// 		"success", func() {}, true,
	// 	},
	// 	{
	// 		"consensus state not found", func() {
	// 			// use height with no consensus state set
	// 			height = height.(clienttypes.Height).Increment()
	// 		}, false,
	// 	},
	// 	{
	// 		"not a consensus state interface", func() {
	// 			// marshal an empty client state and set as consensus state
	// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
	// 			clientStateBz := suite.chainA.App.GetIBCKeeper().ClientKeeper.MustMarshalClientState(&types.ClientState{})
	// 			store.Set(host.ConsensusStateKey(height), clientStateBz)
	// 		}, false,
	// 	},
	// 	{
	// 		"invalid consensus state (solomachine)", func() {
	// 			// marshal and set solomachine consensus state
	// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
	// 			consensusStateBz := suite.chainA.App.GetIBCKeeper().ClientKeeper.MustMarshalConsensusState(&solomachinetypes.ConsensusState{})
	// 			store.Set(host.ConsensusStateKey(height), consensusStateBz)
	// 		}, false,
	// 	},
	// }

	// for _, tc := range testCases {
	// 	tc := tc

	// 	suite.Run(tc.name, func() {
	// 		suite.SetupTest()
	// 		path = ibctesting.NewPath(suite.chainA, suite.chainB)

	// 		suite.coordinator.Setup(path)
	// 		clientState := suite.chainA.GetClientState(path.EndpointA.ClientID)
	// 		height = clientState.GetLatestHeight()

	// 		tc.malleate() // change vars as necessary

	// 		store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
	// 		consensusState, err := tendermint.GetConsensusState(store, suite.chainA.Codec, height)

	// 		if tc.expPass {
	// 			suite.Require().NoError(err)
	// 			expConsensusState, found := suite.chainA.GetConsensusState(path.EndpointA.ClientID, height)
	// 			suite.Require().True(found)
	// 			suite.Require().Equal(expConsensusState, consensusState)
	// 		} else {
	// 			suite.Require().Error(err)
	// 			suite.Require().Nil(consensusState)
	// 		}
	// 	})
	// }
}

func (suite *GrandpaTestSuite) TestGetProcessedTime() {
	// setup
	path := ibctesting.NewPath(suite.chainA, suite.chainB)

	suite.coordinator.UpdateTime()
	// coordinator increments time before creating client
	expectedTime := suite.chainA.CurrentHeader.Time.Add(ibctesting.TimeIncrement)

	// Verify ProcessedTime on CreateClient
	err := path.EndpointA.CreateClient()
	suite.Require().NoError(err)

	clientState := suite.chainA.GetClientState(path.EndpointA.ClientID)
	height := clientState.GetLatestHeight()

	store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
	actualTime, ok := tendermint.GetProcessedTime(store, height)
	suite.Require().True(ok, "could not retrieve processed time for stored consensus state")
	suite.Require().Equal(uint64(expectedTime.UnixNano()), actualTime, "retrieved processed time is not expected value")

	suite.coordinator.UpdateTime()
	// coordinator increments time before updating client
	expectedTime = suite.chainA.CurrentHeader.Time.Add(ibctesting.TimeIncrement)

	// Verify ProcessedTime on UpdateClient
	err = path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	clientState = suite.chainA.GetClientState(path.EndpointA.ClientID)
	height = clientState.GetLatestHeight()

	store = suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
	actualTime, ok = tendermint.GetProcessedTime(store, height)
	suite.Require().True(ok, "could not retrieve processed time for stored consensus state")
	suite.Require().Equal(uint64(expectedTime.UnixNano()), actualTime, "retrieved processed time is not expected value")

	// try to get processed time for height that doesn't exist in store
	_, ok = tendermint.GetProcessedTime(store, clienttypes.NewHeight(1, 1))
	suite.Require().False(ok, "retrieved processed time for a non-existent consensus state")
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
	nextValsHash := []byte("nextVals")

	// Set iteration keys and consensus states
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), clienttypes.NewHeight(0, 1))
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", clienttypes.NewHeight(0, 1), types.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("hash0-1")), nextValsHash))
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), clienttypes.NewHeight(4, 9))
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", clienttypes.NewHeight(4, 9), types.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("hash4-9")), nextValsHash))
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), clienttypes.NewHeight(0, 10))
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", clienttypes.NewHeight(0, 10), types.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("hash0-10")), nextValsHash))
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), clienttypes.NewHeight(0, 4))
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", clienttypes.NewHeight(0, 4), types.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("hash0-4")), nextValsHash))
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), clienttypes.NewHeight(40, 1))
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", clienttypes.NewHeight(40, 1), types.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("hash40-1")), nextValsHash))

	var testArr []string
	cb := func(height exported.Height) bool {
		testArr = append(testArr, height.String())
		return false
	}

	tendermint.IterateConsensusStateAscending(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), cb)
	expectedArr := []string{"0-1", "0-4", "0-10", "4-9", "40-1"}
	suite.Require().Equal(expectedArr, testArr)
}

func (suite *GrandpaTestSuite) TestGetNeighboringConsensusStates() {
	nextValsHash := []byte("nextVals")
	cs01 := tendermint.NewConsensusState(time.Now().UTC(), commitmenttypes.NewMerkleRoot([]byte("hash0-1")), nextValsHash)
	cs04 := tendermint.NewConsensusState(time.Now().UTC(), commitmenttypes.NewMerkleRoot([]byte("hash0-4")), nextValsHash)
	cs49 := tendermint.NewConsensusState(time.Now().UTC(), commitmenttypes.NewMerkleRoot([]byte("hash4-9")), nextValsHash)
	height01 := clienttypes.NewHeight(0, 1)
	height04 := clienttypes.NewHeight(0, 4)
	height49 := clienttypes.NewHeight(4, 9)

	// Set iteration keys and consensus states
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), height01)
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", height01, cs01)
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), height04)
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", height04, cs04)
	tendermint.SetIterationKey(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), height49)
	suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), "testClient", height49, cs49)

	prevCs01, ok := tendermint.GetPreviousConsensusState(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), suite.chainA.Codec, height01)
	suite.Require().Nil(prevCs01, "consensus state exists before lowest consensus state")
	suite.Require().False(ok)
	prevCs49, ok := tendermint.GetPreviousConsensusState(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), suite.chainA.Codec, height49)
	suite.Require().Equal(cs04, prevCs49, "previous consensus state is not returned correctly")
	suite.Require().True(ok)

	nextCs01, ok := tendermint.GetNextConsensusState(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), suite.chainA.Codec, height01)
	suite.Require().Equal(cs04, nextCs01, "next consensus state not returned correctly")
	suite.Require().True(ok)
	nextCs49, ok := tendermint.GetNextConsensusState(suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), "testClient"), suite.chainA.Codec, height49)
	suite.Require().Nil(nextCs49, "next consensus state exists after highest consensus state")
	suite.Require().False(ok)
}
