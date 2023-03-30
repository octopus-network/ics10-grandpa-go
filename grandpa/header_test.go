package grandpa_test

import (
	"time"

	ibcgptypes "github.com/octopus-network/ics10-grandpa-go/grandpa"
)

func (suite *GrandpaTestSuite) TestGetHeight() {
	// header := suite.chainA.LastHeader
	height := gpHeader.GetHeight()
	suite.Suite.T().Logf("gpHeader.GetHeight(): %+v", gpHeader.GetHeight())
	suite.Require().NotEqual(uint64(0), height)
}

func (suite *GrandpaTestSuite) TestGetTime() {
	// header := suite.chainA.LastHeader
	headerTime := gpHeader.GetTime()
	suite.Suite.T().Logf("gpHeader.GetTime(): %+v", headerTime)
	suite.Require().NotEqual(time.Time{}, headerTime)
}

func (suite *GrandpaTestSuite) TestHeaderValidateBasic() {
	var header *ibcgptypes.Header
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"valid header", func() {
			header = &gpHeader
		}, true},
		{"header is nil", func() {
			header = &ibcgptypes.Header{}
		}, false},
		{"beefy mmr is nil", func() {
			header = &ibcgptypes.Header{
				BeefyMmr: ibcgptypes.BeefyMMR{},
				Message:  gpHeader.Message,
			}
		}, false},
		{"header message is nil", func() {
			header = &ibcgptypes.Header{
				BeefyMmr: gpHeader.BeefyMmr,
				Message:  nil,
			}
		}, false},
	}

	suite.Require().Equal(ibcgptypes.ModuleName, gpHeader.ClientType())

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.Suite.T().Logf("tc.name: %s", tc.name)
			suite.SetupTest()

			tc.malleate()
			err := header.ValidateBasic()
			suite.Suite.T().Logf("header.ValidateBasic() err: %+v", err)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
