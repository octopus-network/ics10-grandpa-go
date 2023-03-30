package grandpa_test

import (
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcgptypes "github.com/octopus-network/ics10-grandpa-go/grandpa"
)

func (suite *GrandpaTestSuite) TestValidate() {
	var clientState *ibcgptypes.ClientState
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"valid client",
			func() {
				clientState = &gpClientState
			}, true,
		},
		{
			"beefy height must > zero",
			func() {
				gpClientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), 0)
				clientState = &gpClientState
			}, false,
		},
	}

	for i, tc := range testCases {
		tc.malleate()
		err := clientState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Suite.T().Logf("clientState.Validate() err: %+v", err)
		}
	}
}
