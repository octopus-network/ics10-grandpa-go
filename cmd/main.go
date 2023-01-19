package main

// import (
// 	"os"

// 	"github.com/cosmos/cosmos-sdk/server"
// 	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

// 	"github.com/cosmos/ibc-go/v6/testing/simapp"
// 	"github.com/cosmos/ibc-go/v6/testing/simapp/simd/cmd"
// )

// func main() {
// 	rootCmd, _ := cmd.NewRootCmd()

// 	if err := svrcmd.Execute(rootCmd, "simd", simapp.DefaultNodeHome); err != nil {
// 		switch e := err.(type) {
// 		case server.ErrorCode:
// 			os.Exit(e.Code)

// 		default:
// 			os.Exit(1)
// 		}
// 	}
// }

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
)

var (
	BEEFY_TEST_MODE    = os.Getenv("BEEFY_TEST_MODE")
	RPC_CLIENT_ADDRESS = os.Getenv("RPC_CLIENT_ADDRESS")
	UPDATE_STATE_MODE  = os.Getenv("UPDATE_STATE_MODE")
)

// var endpoint = "wss://rococo-rpc.polkadot.io"

// var relay_chain_endpoint = "wss://rpc.polkadot.io"
var relay_chain_endpoint = "ws://127.0.0.1:9944"

const PARA_ID uint32 = 2000

func TestSimpleConnect(t *testing.T) {
	// The following example shows how to instantiate a Substrate API and use it to connect to a node
	// rococo_endpoint := "wss://rococo-rpc.polkadot.io"

	// api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	api, err := gsrpc.NewSubstrateAPI(relay_chain_endpoint)
	if err != nil {
		panic(err)
	}

	chain, err := api.RPC.System.Chain()
	if err != nil {
		panic(err)
	}
	t.Log("chain:", chain)
	nodeName, err := api.RPC.System.Name()
	if err != nil {
		panic(err)
	}
	t.Log("nodeName:", nodeName)
	nodeVersion, err := api.RPC.System.Version()
	t.Log("nodeVersion:", nodeVersion)
	if err != nil {
		panic(err)
	}

	fmt.Printf("You are connected to chain %v using %v v%v\n", chain, nodeName, nodeVersion)
}

func TestSubNewBlock(t *testing.T) {
	// This example shows how to subscribe to new blocks.
	//
	// It displays the block number every time a new block is seen by the node you are connected to.
	//
	// NOTE: The example runs until 10 blocks are received or until you stop it with CTRL+C

	// api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	api, err := gsrpc.NewSubstrateAPI(relay_chain_endpoint)
	if err != nil {
		panic(err)
	}

	sub, err := api.RPC.Chain.SubscribeNewHeads()
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	count := 0

	for {
		head := <-sub.Chan()
		t.Log("Chain is at block:", head.Number)

		count++

		if count == 10 {
			sub.Unsubscribe()
			break
		}
	}
}

// func main() {

// 	// api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)

// 	api, err := gsrpc.NewSubstrateAPI(rococo_endpoint)
// 	if err != nil {
// 		fmt.Printf("connection err,%s", err)

// 	}
// 	// block_hash, err := api.RPC.Beefy.GetFinalizedHead()

// 	// if err != nil {
// 	// 	fmt.Printf("GetFinalizedHead err,%s", err)

// 	// }

// 	// fmt.Printf("beefy GetFinalizedHead: %#v\n", block_hash)

// 	sub, err := api.RPC.Beefy.SubscribeJustifications()
// 	if err != nil {
// 		fmt.Printf("SubscribeJustifications err,%s", err)

// 	}

// 	defer sub.Unsubscribe()

// 	timeout := time.After(300 * time.Second)
// 	received := 0

// 	for {
// 		select {
// 		case commitment := <-sub.Chan():
// 			fmt.Printf("%#v\n", commitment)
// 			// fmt.Printf("commitment:", commitment)
// 			received++

// 			if received >= 2 {
// 				return
// 			}
// 		case <-timeout:
// 			fmt.Printf("timeout reached without getting 2 notifications from subscription")
// 			return
// 		}
// 	}
// }

// A [SignedCommitment] with a version number.
//
// This variant will be appended to the block justifications for the block
// for which the signed commitment has been generated.
//
// Note that this enum is subject to change in the future with introduction
// of additional cryptographic primitives to BEEFY.
// refer to :https://github.com/paritytech/substrate/blob/b85246bf1156f9f58825f0be48e6086128cc0bbd/primitives/beefy/src/commitment.rs#L237
type VersionedFinalityProof struct {
	Version          uint8
	SignedCommitment types.SignedCommitment
}

func (vfp *VersionedFinalityProof) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 1:
		vfp.Version = 1
		err = decoder.Decode(&vfp.SignedCommitment)

	}

	if err != nil {
		return err
	}

	return nil
}

func (vfp VersionedFinalityProof) Encode(encoder scale.Encoder) error {
	var err1, err2 error

	// if v.V1.(types.SignedCommitment) {
	// 	err1 = encoder.PushByte(1)
	// 	err2 = encoder.Encode(v.V1)
	// }
	switch vfp.Version {
	case 1:
		err1 = encoder.PushByte(1)
		err2 = encoder.Encode(vfp.SignedCommitment)

	}

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func main() {

	api, err := gsrpc.NewSubstrateAPI(relay_chain_endpoint)
	if err != nil {
		// fmt.Printf("connection err,%s", err)
		log.Printf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := api.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	if err != nil && err.Error() == "Method not found" {
		fmt.Printf("skipping since beefy module is not available")
		// log.Printf("skipping since beefy module is not available %v", err)
	}

	// fmt.Printf("subscribed to %s\n", polkadot_endpoint)
	log.Printf("subscribed to %s\n", relay_chain_endpoint)
	// assert.NoError(t, err)
	defer sub.Unsubscribe()

	timeout := time.After(24 * time.Hour)
	received := 0

	for {
		select {
		case msg := <-ch:
			log.Printf("encoded msg: %s\n", msg)

			// s := &types.SignedCommitment{}
			s := &VersionedFinalityProof{}
			err := codec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			log.Printf("encoded msg: %#v\n", s)

			received++

			if received >= 1800 {
				return
			}
		case <-timeout:
			log.Printf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}
