package util

import (
	"encoding/binary"
	"fmt"
	"log"

	"encoding/json"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/client"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/go-substrate-rpc-client/v4/xxhash"
	"github.com/ethereum/go-ethereum/crypto"
	grandpa "github.com/octopus-network/ics10-grandpa-go/types"
)

func GetBeefyAuthorities(blockHash types.Hash, api *gsrpc.SubstrateAPI, method string) ([][]byte, error) {
	// blockHash, err := conn.RPC.Chain.GetBlockHash(uint64(blockNumber))
	// if err != nil {
	// 	return nil, err
	// }

	// Fetch metadata
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	storageKey, err := types.CreateStorageKey(meta, "Beefy", method, nil, nil)
	if err != nil {
		return nil, err
	}

	var authorities grandpa.Authorities

	ok, err := api.RPC.State.GetStorage(storageKey, &authorities, blockHash)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("beefy authorities not found")
	}
	log.Printf("authority count: %d\n", len(authorities))
	// Convert from ecdsa public key to ethereum address
	var authorityEthereumAddresses [][]byte
	for _, authority := range authorities {
		// log.Printf("authority pubkey: %s\n", codec.HexEncodeToString(authority[:]))
		pub, err := crypto.DecompressPubkey(authority[:])
		if err != nil {
			return nil, err
		}
		ethereumAddress := crypto.PubkeyToAddress(*pub)
		// log.Printf("authority ethereumAddress: %s\n", ethereumAddress)
		authorityEthereumAddresses = append(authorityEthereumAddresses, ethereumAddress[:])
	}

	return authorityEthereumAddresses, nil
}

func GetBeefyAuthoritySet(blockHash types.Hash, api *gsrpc.SubstrateAPI, method string) (types.BeefyNextAuthoritySet, error) {
	// blockHash, err := conn.RPC.Chain.GetBlockHash(uint64(blockNumber))
	// if err != nil {
	// 	return nil, err
	// }
	var authoritySet types.BeefyNextAuthoritySet
	// Fetch metadata
	log.Printf("blockHash: %#v\n", codec.HexEncodeToString(blockHash[:]))
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return authoritySet, err
	}

	storageKey, err := types.CreateStorageKey(meta, "MmrLeaf", method, nil, nil)
	if err != nil {
		return authoritySet, err
	}
	log.Printf("storageKey: %#v\n", codec.HexEncodeToString(storageKey[:]))

	// var authoritySet *grandpa.BeefyAuthoritySet
	// var authoritySetBytes []byte

	ok, err := api.RPC.State.GetStorage(storageKey, &authoritySet, blockHash)
	// raw, err := api.RPC.State.GetStorageRaw(storageKey, blockHash)

	if err != nil {
		log.Printf("get storage err: %#s\n", err)
		return authoritySet, err
	}

	if !ok {
		return authoritySet, fmt.Errorf("beefy authority set not found")
	}
	log.Printf("BeefyAuthoritySet on chain : %#v\n", authoritySet)

	return authoritySet, nil

}

func FetchParachainHeader(conn *gsrpc.SubstrateAPI, paraId uint32, blockHash types.Hash) ([]byte, error) {
	// Fetch metadata
	meta, err := conn.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	paraIdEncoded := make([]byte, 4)
	binary.LittleEndian.PutUint32(paraIdEncoded, paraId)

	storageKey, err := types.CreateStorageKey(meta, "Paras", "Heads", paraIdEncoded)

	if err != nil {
		return nil, err
	}

	var parachainHeaders []byte

	ok, err := conn.RPC.State.GetStorage(storageKey, &parachainHeaders, blockHash)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("parachain header not found")
	}

	return parachainHeaders, nil
}

func FetchParaIDs(conn *gsrpc.SubstrateAPI, blockHash types.Hash) ([]uint32, error) {
	// Fetch metadata
	meta, err := conn.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	storageKey, err := types.CreateStorageKey(meta, "Paras", "Parachains", nil, nil)
	if err != nil {
		return nil, err
	}

	var paraIds []uint32

	ok, err := conn.RPC.State.GetStorage(storageKey, &paraIds, blockHash)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("beefy authorities not found")
	}

	return paraIds, nil
}

// CreateStorageKeyPrefix creates a key prefix for keys of a map.
// Can be used as an input to the state.GetKeys() RPC, in order to list the keys of map.
func CreateStorageKeyPrefix(prefix, method string) []byte {
	return append(xxhash.New128([]byte(prefix)).Sum(nil), xxhash.New128([]byte(method)).Sum(nil)...)
}

// GenerateMmrBatchProofResponse contains the generate batch proof rpc response
type GenerateMmrBatchProofResponse struct {
	BlockHash types.H256
	Leaves    []types.MMRLeaf
	Proof     MmrBatchProof
}

// UnmarshalJSON fills u with the JSON encoded byte array given by b
func (d *GenerateMmrBatchProofResponse) UnmarshalJSON(bz []byte) error {
	var tmp struct {
		BlockHash string `json:"blockHash"`
		Leaves    string `json:"leaves"`
		Proof     string `json:"proof"`
	}
	if err := json.Unmarshal(bz, &tmp); err != nil {
		return err
	}
	err := codec.DecodeFromHex(tmp.BlockHash, &d.BlockHash)
	if err != nil {
		return err
	}

	var opaqueLeaves [][]byte
	err = codec.DecodeFromHex(tmp.Leaves, &opaqueLeaves)
	if err != nil {
		return err
	}
	for _, leaf := range opaqueLeaves {

		var mmrLeaf types.MMRLeaf
		err := codec.Decode(leaf, &mmrLeaf)
		if err != nil {
			return err
		}
		d.Leaves = append(d.Leaves, mmrLeaf)
	}
	err = codec.DecodeFromHex(tmp.Proof, &d.Proof)
	if err != nil {
		return err
	}
	return nil
}

// MmrProof is a MMR proof
type MmrBatchProof struct {
	// The index of the leaf the proof is for.
	LeafIndex []types.U64
	// Number of leaves in MMR, when the proof was generated.
	LeafCount types.U64
	// Proof elements (hashes of siblings of inner nodes on the path to the leaf).
	Items []types.H256
}

// GenerateProof retrieves a MMR proof and leaf for the specified leave index, at the given blockHash (useful to query a
// proof at an earlier block, likely with another MMR root)
func GenerateBatchProof(conn *gsrpc.SubstrateAPI, blockHash *types.Hash, indices []uint64) (GenerateMmrBatchProofResponse, error) {
	var res GenerateMmrBatchProofResponse
	err := client.CallWithBlockHash(conn.Client, &res, "mmr_generateBatchProof", blockHash, indices)
	if err != nil {
		return GenerateMmrBatchProofResponse{}, err
	}

	return res, nil
}

func QueryParaChainStorage(conn *gsrpc.SubstrateAPI, paraID uint32, startBlockNumber uint32, endBlockNumber uint32) ([]types.StorageChangeSet, error) {

	log.Printf("startBlockNumber: %d", startBlockNumber)
	log.Printf("endBlockNumber: %d", endBlockNumber)
	log.Printf("parachian id : %d", paraID)
	var paraHeaderKeys []types.StorageKey

	// create full storage key for our own paraId
	keyPrefix := CreateStorageKeyPrefix("Paras", "Heads")
	log.Printf("keyPrefix: %s", codec.HexEncodeToString(keyPrefix[:]))
	encodedParaID, err := codec.Encode(paraID)
	log.Printf("encodedParaID: %s", codec.HexEncodeToString(encodedParaID[:]))
	if err != nil {
		return []types.StorageChangeSet{}, err
	}

	twoXHash := xxhash.New64(encodedParaID).Sum(nil)
	log.Printf("encodedParaID twoXHash: %s", codec.HexEncodeToString(twoXHash[:]))
	// full key path in the storage source: https://www.shawntabrizi.com/assets/presentations/substrate-storage-deep-dive.pdf
	// xx128("Paras") + xx128("Heads") + xx64(Encode(paraId)) + Encode(paraId)
	fullKey := append(append(keyPrefix, twoXHash[:]...), encodedParaID...)
	log.Printf("fullKey: %s", codec.HexEncodeToString(fullKey[:]))
	paraHeaderKeys = append(paraHeaderKeys, fullKey)

	var changSet []types.StorageChangeSet

	for i := startBlockNumber; i <= endBlockNumber; i++ {
		blockHash, err := conn.RPC.Chain.GetBlockHash(uint64(i))
		if err != nil {
			return []types.StorageChangeSet{}, err
		}
		log.Printf("blockHash: %s\n", codec.HexEncodeToString(blockHash[:]))

		cs, err := conn.RPC.State.QueryStorageAt(paraHeaderKeys, blockHash)
		if err != nil {
			return []types.StorageChangeSet{}, err
		}
		log.Printf("cs: %#v", cs)

		changSet = append(changSet, cs...)
		// log.Printf("changeSet: %#v", changeSet)

	}
	log.Printf("changeSet: %#v", changSet)

	return changSet, nil
}
