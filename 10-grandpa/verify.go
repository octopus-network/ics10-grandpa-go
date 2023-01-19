package grandpa

import (
	"log"

	"github.com/ComposableFi/go-merkle-trees/hasher"
	"github.com/ComposableFi/go-merkle-trees/merkle"
	merkletypes "github.com/ComposableFi/go-merkle-trees/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/ethereum/go-ethereum/crypto"
)

type SignatureWithIdx struct {
	// actual signature bytes
	Signature []byte
	// authority leaf index in the merkle tree.
	AuthorityIndex uint32
}
type ConvertedSignedCommitment struct {
	Commitment types.Commitment
	Signatures []SignatureWithIdx
}

type Authorities = [][33]uint8

// type BeefyAuthoritySet struct {
// 	// ID
// 	Id uint64
// 	// Number of validators in the set.
// 	Len uint32
// 	// Merkle Root Hash build from BEEFY uncompressed AuthorityIds.
// 	Root [32]byte
// }

// create authority proof
func CreateAuthorityProof(sc types.SignedCommitment, authorityTree merkle.Tree) (ConvertedSignedCommitment, [][]byte, error) {
	var sigIdxes []SignatureWithIdx
	var authorityIndices []uint64
	// luckily for us, this is already sorted and maps to the right authority index in the authority root.
	for i, v := range sc.Signatures {
		if v.IsSome() {
			_, sig := v.Unwrap()
			sigIdxes = append(sigIdxes, SignatureWithIdx{
				Signature:      sig[:],
				AuthorityIndex: uint32(i),
			})
			log.Printf("authority signatures: %#v\n", sigIdxes)
			authorityIndices = append(authorityIndices, uint64(i))
		}
	}
	authoritiesProof := authorityTree.Proof(authorityIndices).ProofHashes()
	var csc = ConvertedSignedCommitment{
		Commitment: sc.Commitment,
		Signatures: sigIdxes,
	}
	log.Printf("authority proofs: %#v\n", authoritiesProof)
	return csc, authoritiesProof, nil
}

// verify authority signatures
func VerifyAuthoritySignatures(csc ConvertedSignedCommitment, bas BeefyAuthoritySet, proofHashes [][]byte, merkleRoot SizedByte32) error {

	// beefy authorities are signing the hash of the scale-encoded Commitment
	commitmentBytes, err := codec.Encode(&csc.Commitment)
	if err != nil {
		return err
	}

	// take keccak hash of the commitment scale-encoded
	commitmentHash := crypto.Keccak256(commitmentBytes)

	// array of leaves in the authority merkle root.
	var authorityLeaves []merkletypes.Leaf

	for i := 0; i < len(csc.Signatures); i++ {
		signature := csc.Signatures[i]
		// recover uncompressed public key from signature
		pubkey, err := crypto.SigToPub(commitmentHash, signature.Signature)
		if err != nil {
			return err
		}

		// convert public key to ethereum address.
		address := crypto.PubkeyToAddress(*pubkey)
		authorityLeaf := merkletypes.Leaf{
			Hash:  crypto.Keccak256(address[:]),
			Index: uint64(signature.AuthorityIndex),
		}
		authorityLeaves = append(authorityLeaves, authorityLeaf)
	}
	authoritiesProof := merkle.NewProof(authorityLeaves, proofHashes, uint64(bas.Len), hasher.Keccak256Hasher{})
	calMerkleRoot, err := authoritiesProof.RootHex()
	if err != nil {
		return err
	}
	log.Printf("cal merkle root: %s\n", calMerkleRoot)
	log.Printf("expected merkle root: %s\n", codec.HexEncodeToString(merkleRoot[:]))

	valid, err := authoritiesProof.Verify(merkleRoot[:])
	if err != nil || !valid {
		return err
	}

	return nil
}
