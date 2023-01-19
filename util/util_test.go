package util_test

import (
	"log"
	"testing"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/go-substrate-rpc-client/v4/xxhash"
	"github.com/octopus-network/grandpa-go/util"
	"github.com/stretchr/testify/require"
)

const PARA_ID uint32 = 2087
const RPC_CLIENT_ADDRESS = "wss://rococo-rpc.polkadot.io"


func TestValidatorSetCodec(t *testing.T) {

	var validatorSet1 = types.BeefyNextAuthoritySet{
		ID:   6222,
		Len:  83,
		Root: [32]byte{242, 71, 234, 49, 93, 55, 186, 220, 142, 244, 51, 94, 85, 241, 146, 62, 213, 162, 250, 37, 110, 101, 244, 99, 128, 6, 194, 124, 44, 64, 44, 140},
	}
	encodeData, err := codec.Encode(validatorSet1)
	require.NoError(t, err)
	log.Printf("encoded validatorSet : %#v\n", codec.HexEncodeToString(encodeData[:]))

	var validatorSet2 types.BeefyNextAuthoritySet

	err = codec.Decode(encodeData, &validatorSet2)

	if err != nil {
		log.Printf("decode err: %#s\n", err)
	}
	log.Printf("decoded validatorSet: %#v\n", validatorSet2)

}

func TestQueryStorage(t *testing.T) {
	relayApi, err := gsrpc.NewSubstrateAPI(RPC_CLIENT_ADDRESS)
	require.NoError(t, err)

	var startBlockNumber = 3707562
	log.Printf("startBlockNumber: %d", startBlockNumber)
	log.Printf("startBlockNumber+1: %d", startBlockNumber+1)
	startFinalizedHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(startBlockNumber + 1))
	require.NoError(t, err)
	log.Printf("startFinalizedHash: %s\n", codec.HexEncodeToString(startFinalizedHash[:]))

	var endBlockNumber = 3707570
	log.Printf("endBlockNumber: %d", endBlockNumber)
	endFinalizedHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(endBlockNumber))
	require.NoError(t, err)
	log.Printf("endBlockNumber: %s\n", codec.HexEncodeToString(endFinalizedHash[:]))

	var paraHeaderKeys []types.StorageKey

	// create full storage key for our own paraId
	keyPrefix := util.CreateStorageKeyPrefix("Paras", "Heads")
	log.Printf("keyPrefix: %s", codec.HexEncodeToString(keyPrefix[:]))
	// so we can query all blocks from lastfinalized to latestBeefyHeight
	log.Printf("PARA_ID: %d", PARA_ID)
	encodedParaID, err := codec.Encode(PARA_ID)
	log.Printf("encodedParaID: %s", codec.HexEncodeToString(encodedParaID[:]))
	require.NoError(t, err)

	twoXHash := xxhash.New64(encodedParaID).Sum(nil)
	log.Printf("encodedParaID twoXHash: %s", codec.HexEncodeToString(twoXHash[:]))
	// full key path in the storage source: https://www.shawntabrizi.com/assets/presentations/substrate-storage-deep-dive.pdf
	// xx128("Paras") + xx128("Heads") + xx64(Encode(paraId)) + Encode(paraId)
	fullKey := append(append(keyPrefix, twoXHash[:]...), encodedParaID...)
	log.Printf("fullKey: %s", codec.HexEncodeToString(fullKey[:]))
	paraHeaderKeys = append(paraHeaderKeys, fullKey)

	changeSet, err := relayApi.RPC.State.QueryStorage(paraHeaderKeys, startFinalizedHash, endFinalizedHash)
	require.NoError(t, err)
	log.Printf("changeSet: %#v", changeSet)
}

func TestQueryStorageAt(t *testing.T) {
	relayApi, err := gsrpc.NewSubstrateAPI(RPC_CLIENT_ADDRESS)
	require.NoError(t, err)

	var startBlockNumber = 3707562
	log.Printf("startBlockNumber: %d", startBlockNumber)
	log.Printf("startBlockNumber+1: %d", startBlockNumber+1)
	startFinalizedHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(startBlockNumber + 1))
	require.NoError(t, err)
	log.Printf("startFinalizedHash: %s\n", codec.HexEncodeToString(startFinalizedHash[:]))

	var endBlockNumber = 3707570
	log.Printf("endBlockNumber: %d", endBlockNumber)
	endFinalizedHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(endBlockNumber))
	require.NoError(t, err)
	log.Printf("endBlockNumber: %s\n", codec.HexEncodeToString(endFinalizedHash[:]))

	var paraHeaderKeys []types.StorageKey

	// create full storage key for our own paraId
	keyPrefix := util.CreateStorageKeyPrefix("Paras", "Heads")
	log.Printf("keyPrefix: %s", codec.HexEncodeToString(keyPrefix[:]))
	// so we can query all blocks from lastfinalized to latestBeefyHeight
	log.Printf("PARA_ID: %d", PARA_ID)
	encodedParaID, err := codec.Encode(PARA_ID)
	log.Printf("encodedParaID: %s", codec.HexEncodeToString(encodedParaID[:]))
	require.NoError(t, err)

	twoXHash := xxhash.New64(encodedParaID).Sum(nil)
	log.Printf("encodedParaID twoXHash: %s", codec.HexEncodeToString(twoXHash[:]))
	// full key path in the storage source: https://www.shawntabrizi.com/assets/presentations/substrate-storage-deep-dive.pdf
	// xx128("Paras") + xx128("Heads") + xx64(Encode(paraId)) + Encode(paraId)
	fullKey := append(append(keyPrefix, twoXHash[:]...), encodedParaID...)
	log.Printf("fullKey: %s", codec.HexEncodeToString(fullKey[:]))
	paraHeaderKeys = append(paraHeaderKeys, fullKey)

	var changSet []types.StorageChangeSet

	for i := startBlockNumber + 1; i <= endBlockNumber; i++ {
		blockHash, err := relayApi.RPC.Chain.GetBlockHash(uint64(i))
		log.Printf("blockHash: %s\n", codec.HexEncodeToString(blockHash[:]))
		require.NoError(t, err)
		cs, err := relayApi.RPC.State.QueryStorageAt(paraHeaderKeys, blockHash)
		require.NoError(t, err)
		log.Printf("cs: %#v", cs)
		changSet = append(changSet, cs...)
		// log.Printf("changeSet: %#v", changeSet)

	}
	log.Printf("changeSet: %#v", changSet)
}
