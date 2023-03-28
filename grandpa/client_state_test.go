package grandpa_test

// import (
// 	"time"

// 	"github.com/cosmos/ibc-go/modules/light-clients/07-tendermint/types"
// 	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
// 	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
// 	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
// 	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
// 	"github.com/cosmos/ibc-go/v7/modules/core/exported"
// 	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
// 	ibctesting "github.com/cosmos/ibc-go/v7/testing"
// 	ibcmock "github.com/cosmos/ibc-go/v7/testing/mock"
// 	ibcgptypes "github.com/octopus-network/ics10-grandpa-go/grandpa"
// )

// const (
// 	testClientID     = "clientidone"
// 	testConnectionID = "connectionid"
// 	testPortID       = "testportid"
// 	testChannelID    = "testchannelid"
// 	testSequence     = 1

// 	// Do not change the length of these variables
// 	fiftyCharChainID    = "12345678901234567890123456789012345678901234567890"
// 	fiftyOneCharChainID = "123456789012345678901234567890123456789012345678901"
// )

// var invalidProof = []byte("invalid proof")

// func (suite *GrandpaTestSuite) TestStatus() {
// 	var (
// 		path        *ibctesting.Path
// 		clientState *tendermint.ClientState
// 	)

// 	testCases := []struct {
// 		name      string
// 		malleate  func()
// 		expStatus exported.Status
// 	}{
// 		{"client is active", func() {}, exported.Active},
// 		{"client is frozen", func() {
// 			clientState.FrozenHeight = clienttypes.NewHeight(0, 1)
// 			path.EndpointA.SetClientState(clientState)
// 		}, exported.Frozen},
// 		{"client status without consensus state", func() {
// 			clientState.LatestHeight = clientState.LatestHeight.Increment().(clienttypes.Height)
// 			path.EndpointA.SetClientState(clientState)
// 		}, exported.Expired},
// 		{"client status is expired", func() {
// 			suite.coordinator.IncrementTimeBy(clientState.TrustingPeriod)
// 		}, exported.Expired},
// 	}

// 	for _, tc := range testCases {
// 		path = ibctesting.NewPath(suite.chainA, suite.chainB)
// 		suite.coordinator.SetupClients(path)

// 		clientStore := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
// 		clientState = path.EndpointA.GetClientState().(tendermint.ClientState)

// 		tc.malleate()

// 		status := clientState.Status(suite.chainA.GetContext(), clientStore, suite.chainA.App.AppCodec())
// 		suite.Require().Equal(tc.expStatus, status)

// 	}

// }

// func (suite *GrandpaTestSuite) TestValidate() {
// 	var clientState *ibcgptypes.ClientState
// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{"valid client",
// 			func() {
// 				clientState = &gpClientState
// 			}, true,
// 		},
// 		{"beefy height must > zero",
// 			func() {
// 				gpClientState.LatestBeefyHeight = 0
// 				clientState = &gpClientState
// 			}, false,
// 		},
// 	}

// 	for i, tc := range testCases {
// 		tc.malleate()
// 		err := clientState.Validate()
// 		if tc.expPass {
// 			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
// 		} else {
// 			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
// 			suite.Suite.T().Logf("clientState.Validate() err: %+v", err)
// 		}
// 	}
// }

// func (suite *GrandpaTestSuite) TestInitialize() {
// 	testCases := []struct {
// 		name           string
// 		consensusState exported.ConsensusState
// 		expPass        bool
// 	}{
// 		{
// 			name:           "valid consensus",
// 			consensusState: &tendermint.ConsensusState{},
// 			expPass:        true,
// 		},
// 		{
// 			name:           "invalid consensus: consensus state is solomachine consensus",
// 			consensusState: ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2).ConsensusState(),
// 			expPass:        false,
// 		},
// 	}

// 	path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 	err := path.EndpointA.CreateClient()
// 	suite.Require().NoError(err)

// 	clientState := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 	store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)

// 	for _, tc := range testCases {
// 		err := clientState.Initialize(suite.chainA.GetContext(), suite.chainA.Codec, store, tc.consensusState)
// 		if tc.expPass {
// 			suite.Require().NoError(err, "valid case returned an error")
// 		} else {
// 			suite.Require().Error(err, "invalid case didn't return an error")
// 		}
// 	}
// }

// func (suite *GrandpaTestSuite) TestVerifyClientConsensusState() {
// 	testCases := []struct {
// 		name           string
// 		clientState    *tendermint.ClientState
// 		consensusState *tendermint.ConsensusState
// 		prefix         commitmenttypes.MerklePrefix
// 		proof          []byte
// 		expPass        bool
// 	}{
// 		// FIXME: uncomment
// 		// {
// 		// 	name:        "successful verification",
// 		// 	clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height,  commitmenttendermint.GetSDKSpecs()),
// 		// 	consensusState: tendermint.ConsensusState{
// 		// 		Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
// 		// 	},
// 		// 	prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
// 		// 	expPass: true,
// 		// },
// 		{
// 			name:        "ApplyPrefix failed",
// 			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
// 			consensusState: &tendermint.ConsensusState{
// 				Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
// 			},
// 			prefix:  commitmenttypes.MerklePrefix{},
// 			expPass: false,
// 		},
// 		{
// 			name:        "latest client height < height",
// 			clientState: types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
// 			consensusState: &types.ConsensusState{
// 				Root: commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
// 			},
// 			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
// 			expPass: false,
// 		},
// 		{
// 			name:        "proof verification failed",
// 			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttendermint.GetSDKSpecs(), upgradePath, false, false),
// 			consensusState: &tendermint.ConsensusState{
// 				Root:               commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()),
// 				NextValidatorsHash: suite.valsHash,
// 			},
// 			prefix:  commitmenttypes.NewMerklePrefix([]byte("ibc")),
// 			proof:   []byte{},
// 			expPass: false,
// 		},
// 	}

// 	for i, tc := range testCases {
// 		tc := tc

// 		err := tc.clientState.VerifyClientConsensusState(
// 			nil, suite.cdc, height, "chainA", tc.clientState.LatestHeight, tc.prefix, tc.proof, tc.consensusState,
// 		)

// 		if tc.expPass {
// 			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
// 		} else {
// 			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
// 		}
// 	}
// }

// // test verification of the connection on chainB being represented in the
// // light client on chainA
// func (suite *GrandpaTestSuite) TestVerifyConnectionState() {
// 	var (
// 		clientState *tendermint.ClientState
// 		proof       []byte
// 		proofHeight exported.Height
// 		prefix      commitmenttypes.MerklePrefix
// 	)

// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"successful verification", func() {}, true,
// 		},
// 		{
// 			"ApplyPrefix failed", func() {
// 				prefix = commitmenttypes.MerklePrefix{}
// 			}, false,
// 		},
// 		{
// 			"latest client height < height", func() {
// 				proofHeight = clientState.LatestHeight.Increment()
// 			}, false,
// 		},
// 		{
// 			"proof verification failed", func() {
// 				proof = invalidProof
// 			}, false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest() // reset

// 			// setup testing conditions
// 			path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 			suite.coordinator.Setup(path)
// 			connection := path.EndpointB.GetConnection()

// 			var ok bool
// 			clientStateI := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 			clientState, ok = clientStateI.(*tendermint.ClientState)
// 			suite.Require().True(ok)

// 			prefix = suite.chainB.GetPrefix()

// 			// make connection proof
// 			connectionKey := host.ConnectionKey(path.EndpointB.ConnectionID)
// 			proof, proofHeight = suite.chainB.QueryProof(connectionKey)

// 			tc.malleate() // make changes as necessary

// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)

// 			err := clientState.VerifyConnectionState(
// 				store, suite.chainA.Codec, proofHeight, &prefix, proof, path.EndpointB.ConnectionID, connection,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

// // test verification of the channel on chainB being represented in the light
// // client on chainA
// func (suite *GrandpaTestSuite) TestVerifyChannelState() {
// 	// var (
// 	// 	clientState tendermint.ClientState
// 	// 	proof       []byte
// 	// 	proofHeight exported.Height
// 	// 	prefix      commitmenttypes.MerklePrefix
// 	// )

// 	// testCases := []struct {
// 	// 	name     string
// 	// 	malleate func()
// 	// 	expPass  bool
// 	// }{
// 	// 	{
// 	// 		"successful verification", func() {}, true,
// 	// 	},
// 	// 	{
// 	// 		"ApplyPrefix failed", func() {
// 	// 			prefix = commitmenttypes.MerklePrefix{}
// 	// 		}, false,
// 	// 	},
// 	// 	{
// 	// 		"latest client height < height", func() {
// 	// 			proofHeight = clientState.LatestHeight.Increment()
// 	// 		}, false,
// 	// 	},
// 	// 	{
// 	// 		"proof verification failed", func() {
// 	// 			proof = invalidProof
// 	// 		}, false,
// 	// 	},
// 	// }

// 	// for _, tc := range testCases {
// 	// 	tc := tc

// 	// 	suite.Run(tc.name, func() {
// 	// 		suite.SetupTest() // reset

// 	// 		// setup testing conditions
// 	// 		path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 	// 		suite.coordinator.Setup(path)
// 	// 		channel := path.EndpointB.GetChannel()

// 	// 		var ok bool
// 	// 		clientStateI := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 	// 		clientState, ok = clientStateI.(tendermint.ClientState)
// 	// 		suite.Require().True(ok)

// 	// 		prefix = suite.chainB.GetPrefix()

// 	// 		// make channel proof
// 	// 		channelKey := host.ChannelKey(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
// 	// 		proof, proofHeight = suite.chainB.QueryProof(channelKey)

// 	// 		tc.malleate() // make changes as necessary

// 	// 		store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)

// 	// 		err := clientState.VerifyChannelState(
// 	// 			store, suite.chainA.Codec, proofHeight, &prefix, proof,
// 	// 			path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channel,
// 	// 		)

// 	// 		if tc.expPass {
// 	// 			suite.Require().NoError(err)
// 	// 		} else {
// 	// 			suite.Require().Error(err)
// 	// 		}
// 	// 	})
// 	// }
// }

// // test verification of the packet commitment on chainB being represented
// // in the light client on chainA. A send from chainB to chainA is simulated.
// func (suite *GrandpaTestSuite) TestVerifyPacketCommitment() {
// 	var (
// 		clientState      tendermint.ClientState
// 		proof            []byte
// 		delayTimePeriod  uint64
// 		delayBlockPeriod uint64
// 		proofHeight      exported.Height
// 		prefix           commitmenttypes.MerklePrefix
// 	)

// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"successful verification", func() {}, true,
// 		},
// 		{
// 			name: "delay time period has passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Second.Nanoseconds())
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay time period has not passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Hour.Nanoseconds())
// 			},
// 			expPass: false,
// 		},
// 		{
// 			name: "delay block period has passed",
// 			malleate: func() {
// 				delayBlockPeriod = 1
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay block period has not passed",
// 			malleate: func() {
// 				delayBlockPeriod = 1000
// 			},
// 			expPass: false,
// 		},

// 		{
// 			"ApplyPrefix failed", func() {
// 				prefix = commitmenttypes.MerklePrefix{}
// 			}, false,
// 		},
// 		{
// 			"latest client height < height", func() {
// 				proofHeight = clientState.LatestHeight.Increment()
// 			}, false,
// 		},
// 		{
// 			"proof verification failed", func() {
// 				proof = invalidProof
// 			}, false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest() // reset

// 			// setup testing conditions
// 			path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 			suite.coordinator.Setup(path)
// 			packet := channeltypes.NewPacket(ibctesting.MockPacketData, 1, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, clienttypes.NewHeight(0, 100), 0)
// 			err := path.EndpointB.SendPacket(packet)
// 			suite.Require().NoError(err)

// 			var ok bool
// 			clientStateI := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 			clientState, ok = clientStateI.(tendermint.ClientState)
// 			suite.Require().True(ok)

// 			prefix = suite.chainB.GetPrefix()

// 			// make packet commitment proof
// 			packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
// 			proof, proofHeight = path.EndpointB.QueryProof(packetKey)

// 			// reset time and block delays to 0, malleate may change to a specific non-zero value.
// 			delayTimePeriod = 0
// 			delayBlockPeriod = 0
// 			tc.malleate() // make changes as necessary

// 			ctx := suite.chainA.GetContext()
// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)

// 			commitment := channeltypes.CommitPacket(suite.chainA.App.GetIBCKeeper().Codec(), packet)
// 			err = clientState.VerifyPacketCommitment(
// 				ctx, store, suite.chainA.Codec, proofHeight, delayTimePeriod, delayBlockPeriod, &prefix, proof,
// 				packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(), commitment,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

// // test verification of the acknowledgement on chainB being represented
// // in the light client on chainA. A send and ack from chainA to chainB
// // is simulated.
// func (suite *GrandpaTestSuite) TestVerifyPacketAcknowledgement() {
// 	var (
// 		clientState      tendermint.ClientState
// 		proof            []byte
// 		delayTimePeriod  uint64
// 		delayBlockPeriod uint64
// 		proofHeight      exported.Height
// 		prefix           commitmenttypes.MerklePrefix
// 	)

// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"successful verification", func() {}, true,
// 		},
// 		{
// 			name: "delay time period has passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Second.Nanoseconds())
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay time period has not passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Hour.Nanoseconds())
// 			},
// 			expPass: false,
// 		},
// 		{
// 			name: "delay block period has passed",
// 			malleate: func() {
// 				delayBlockPeriod = 1
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay block period has not passed",
// 			malleate: func() {
// 				delayBlockPeriod = 10
// 			},
// 			expPass: false,
// 		},

// 		{
// 			"ApplyPrefix failed", func() {
// 				prefix = commitmenttypes.MerklePrefix{}
// 			}, false,
// 		},
// 		{
// 			"latest client height < height", func() {
// 				proofHeight = clientState.LatestHeight.Increment()
// 			}, false,
// 		},
// 		{
// 			"proof verification failed", func() {
// 				proof = invalidProof
// 			}, false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest() // reset

// 			// setup testing conditions
// 			path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 			suite.coordinator.Setup(path)
// 			packet := channeltypes.NewPacket(ibctesting.MockPacketData, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, clienttypes.NewHeight(0, 100), 0)

// 			// send packet
// 			err := path.EndpointA.SendPacket(packet)
// 			suite.Require().NoError(err)

// 			// write receipt and ack
// 			err = path.EndpointB.RecvPacket(packet)
// 			suite.Require().NoError(err)

// 			var ok bool
// 			clientStateI := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 			clientState, ok = clientStateI.(tendermint.ClientState)
// 			suite.Require().True(ok)

// 			prefix = suite.chainB.GetPrefix()

// 			// make packet acknowledgement proof
// 			acknowledgementKey := host.PacketAcknowledgementKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
// 			proof, proofHeight = suite.chainB.QueryProof(acknowledgementKey)

// 			// reset time and block delays to 0, malleate may change to a specific non-zero value.
// 			delayTimePeriod = 0
// 			delayBlockPeriod = 0
// 			tc.malleate() // make changes as necessary

// 			ctx := suite.chainA.GetContext()
// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)

// 			err = clientState.VerifyPacketAcknowledgement(
// 				ctx, store, suite.chainA.Codec, proofHeight, delayTimePeriod, delayBlockPeriod, &prefix, proof,
// 				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(), ibcmock.MockAcknowledgement.Acknowledgement(),
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

// // test verification of the absent acknowledgement on chainB being represented
// // in the light client on chainA. A send from chainB to chainA is simulated, but
// // no receive.
// func (suite *GrandpaTestSuite) TestVerifyPacketReceiptAbsence() {
// 	var (
// 		clientState      tendermint.ClientState
// 		proof            []byte
// 		delayTimePeriod  uint64
// 		delayBlockPeriod uint64
// 		proofHeight      exported.Height
// 		prefix           commitmenttypes.MerklePrefix
// 	)

// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"successful verification", func() {}, true,
// 		},
// 		{
// 			name: "delay time period has passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Second.Nanoseconds())
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay time period has not passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Hour.Nanoseconds())
// 			},
// 			expPass: false,
// 		},
// 		{
// 			name: "delay block period has passed",
// 			malleate: func() {
// 				delayBlockPeriod = 1
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay block period has not passed",
// 			malleate: func() {
// 				delayBlockPeriod = 10
// 			},
// 			expPass: false,
// 		},

// 		{
// 			"ApplyPrefix failed", func() {
// 				prefix = commitmenttypes.MerklePrefix{}
// 			}, false,
// 		},
// 		{
// 			"latest client height < height", func() {
// 				proofHeight = clientState.LatestHeight.Increment()
// 			}, false,
// 		},
// 		{
// 			"proof verification failed", func() {
// 				proof = invalidProof
// 			}, false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest() // reset

// 			// setup testing conditions
// 			path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 			suite.coordinator.Setup(path)
// 			packet := channeltypes.NewPacket(ibctesting.MockPacketData, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, clienttypes.NewHeight(0, 100), 0)

// 			// send packet, but no recv
// 			err := path.EndpointA.SendPacket(packet)
// 			suite.Require().NoError(err)

// 			var ok bool
// 			clientStateI := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 			clientState, ok = clientStateI.(tendermint.ClientState)
// 			suite.Require().True(ok)

// 			prefix = suite.chainB.GetPrefix()

// 			// make packet receipt absence proof
// 			receiptKey := host.PacketReceiptKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
// 			proof, proofHeight = path.EndpointB.QueryProof(receiptKey)

// 			// reset time and block delays to 0, malleate may change to a specific non-zero value.
// 			delayTimePeriod = 0
// 			delayBlockPeriod = 0
// 			tc.malleate() // make changes as necessary

// 			ctx := suite.chainA.GetContext()
// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)

// 			err = clientState.VerifyPacketReceiptAbsence(
// 				ctx, store, suite.chainA.Codec, proofHeight, delayTimePeriod, delayBlockPeriod, &prefix, proof,
// 				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }

// // test verification of the next receive sequence on chainB being represented
// // in the light client on chainA. A send and receive from chainB to chainA is
// // simulated.
// func (suite *GrandpaTestSuite) TestVerifyNextSeqRecv() {
// 	var (
// 		clientState      tendermint.ClientState
// 		proof            []byte
// 		delayTimePeriod  uint64
// 		delayBlockPeriod uint64
// 		proofHeight      exported.Height
// 		prefix           commitmenttypes.MerklePrefix
// 	)

// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"successful verification", func() {}, true,
// 		},
// 		{
// 			name: "delay time period has passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Second.Nanoseconds())
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay time period has not passed",
// 			malleate: func() {
// 				delayTimePeriod = uint64(time.Hour.Nanoseconds())
// 			},
// 			expPass: false,
// 		},
// 		{
// 			name: "delay block period has passed",
// 			malleate: func() {
// 				delayBlockPeriod = 1
// 			},
// 			expPass: true,
// 		},
// 		{
// 			name: "delay block period has not passed",
// 			malleate: func() {
// 				delayBlockPeriod = 10
// 			},
// 			expPass: false,
// 		},

// 		{
// 			"ApplyPrefix failed", func() {
// 				prefix = commitmenttypes.MerklePrefix{}
// 			}, false,
// 		},
// 		{
// 			"latest client height < height", func() {
// 				proofHeight = clientState.LatestHeight.Increment()
// 			}, false,
// 		},
// 		{
// 			"proof verification failed", func() {
// 				proof = invalidProof
// 			}, false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest() // reset

// 			// setup testing conditions
// 			path := ibctesting.NewPath(suite.chainA, suite.chainB)
// 			path.SetChannelOrdered()
// 			suite.coordinator.Setup(path)
// 			packet := channeltypes.NewPacket(ibctesting.MockPacketData, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, clienttypes.NewHeight(0, 100), 0)

// 			// send packet
// 			err := path.EndpointA.SendPacket(packet)
// 			suite.Require().NoError(err)

// 			// next seq recv incremented
// 			err = path.EndpointB.RecvPacket(packet)
// 			suite.Require().NoError(err)

// 			var ok bool
// 			clientStateI := suite.chainA.GetClientState(path.EndpointA.ClientID)
// 			clientState, ok = clientStateI.(tendermint.ClientState)
// 			suite.Require().True(ok)

// 			prefix = suite.chainB.GetPrefix()

// 			// make next seq recv proof
// 			nextSeqRecvKey := host.NextSequenceRecvKey(packet.GetDestPort(), packet.GetDestChannel())
// 			proof, proofHeight = suite.chainB.QueryProof(nextSeqRecvKey)

// 			// reset time and block delays to 0, malleate may change to a specific non-zero value.
// 			delayTimePeriod = 0
// 			delayBlockPeriod = 0
// 			tc.malleate() // make changes as necessary

// 			ctx := suite.chainA.GetContext()
// 			store := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)

// 			err = clientState.VerifyNextSequenceRecv(
// 				ctx, store, suite.chainA.Codec, proofHeight, delayTimePeriod, delayBlockPeriod, &prefix, proof,
// 				packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence()+1,
// 			)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
