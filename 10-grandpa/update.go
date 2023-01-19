package grandpa

import (
	"bytes"
	"encoding/hex"
	fmt "fmt"
	"log"

	"github.com/ChainSafe/log15"
	"github.com/ComposableFi/go-merkle-trees/hasher"
	"github.com/ComposableFi/go-merkle-trees/merkle"
	"github.com/ComposableFi/go-merkle-trees/mmr"
	merkletypes "github.com/ComposableFi/go-merkle-trees/types"
	scalecodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/ethereum/go-ethereum/crypto"
)

// VerifyClientMessage checks if the clientMessage is of type Header or Misbehaviour and verifies the message
func (cs *ClientState) VerifyClientMessage(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	clientMsg exported.ClientMessage,
) error {
	switch msg := clientMsg.(type) {
	case *Header:
		return cs.verifyHeader(ctx, clientStore, cdc, msg)
	default:
		return clienttypes.ErrInvalidClientType
	}
}

// verifyHeader returns an error if:
// - the client or header provided are not parseable
// - the header is invalid
// - header height is less than or equal to the trusted header height
// - header revision is not equal to trusted header revision
// - header valset commit verification fails
// - header timestamp is past the trusting period in relation to the consensus state
// - header timestamp is less than or equal to the consensus state timestamp
func (cs *ClientState) verifyHeader(
	ctx sdk.Context, clientStore sdk.KVStore, cdc codec.BinaryCodec,
	beefyHeader *Header,
) error {
	var (
		mmrUpdateProof   = beefyHeader.MmrUpdateProof
		authoritiesProof = beefyHeader.MmrUpdateProof.AuthoritiesProof
		signedCommitment = beefyHeader.MmrUpdateProof.SignedCommitment
	)
	log.Printf("verify header --> mmrUpdateProof: %#v\n", mmrUpdateProof)
	log.Printf("verify header --> authoritiesProof: %#v\n", authoritiesProof)
	log.Printf("verify header --> signedCommitment: %#v\n", signedCommitment)

	// checking signatures is expensive (667 authorities for kusama),
	// we want to know if these sigs meet the minimum threshold before proceeding
	// and are by a known authority set (the current one, or the next one)
	if authoritiesThreshold(*cs.Authority) > uint32(len(signedCommitment.Signatures)) ||
		authoritiesThreshold(*cs.NextAuthoritySet) > uint32(len(signedCommitment.Signatures)) {
		return ErrCommitmentNotFinal
	}

	if signedCommitment.Commitment.ValidatorSetId != cs.Authority.Id &&
		signedCommitment.Commitment.ValidatorSetId != cs.NextAuthoritySet.Id {
		return ErrAuthoritySetUnknown
	}

	// beefy authorities are signing the hash of the scale-encoded Commitment
	commitmentBytes, err := scalecodec.Encode(&signedCommitment.Commitment)
	if err != nil {
		return sdkerrors.Wrap(err, ErrInvalidCommitment.Error())
	}

	// take keccak hash of the commitment scale-encoded
	commitmentHash := crypto.Keccak256(commitmentBytes)

	// array of leaves in the authority merkle root.
	var authorityLeaves []merkletypes.Leaf

	for i := 0; i < len(signedCommitment.Signatures); i++ {
		signature := signedCommitment.Signatures[i]
		// recover uncompressed public key from signature
		pubkey, err := crypto.SigToPub(commitmentHash, signature.Signature)
		if err != nil {
			return sdkerrors.Wrap(err, ErrInvalidCommitmentSignature.Error())
		}

		// convert public key to ethereum address.
		address := crypto.PubkeyToAddress(*pubkey)
		authorityLeaf := merkletypes.Leaf{
			Hash:  crypto.Keccak256(address[:]),
			Index: uint64(signature.AuthorityIndex),
		}
		authorityLeaves = append(authorityLeaves, authorityLeaf)
	}

	// flag for if the authority set has been updated
	updatedAuthority := false

	// assert that known authorities signed this commitment, only 2 cases because we already
	// made a prior check to assert that authorities are known
	switch signedCommitment.Commitment.ValidatorSetId {
	case cs.Authority.Id:
		// here we construct a merkle proof, and verify that the public keys which produced this signature
		// are part of the current round.
		authoritiesProof := merkle.NewProof(authorityLeaves, authoritiesProof, uint64(cs.Authority.Len), hasher.Keccak256Hasher{})
		valid, err := authoritiesProof.Verify(cs.Authority.AuthorityRoot[:])
		if err != nil || !valid {
			return sdkerrors.Wrap(err, ErrAuthoritySetUnknown.Error())
		}
		log.Printf("verify authority successfully !\n")

	// new authority set has kicked in
	case cs.NextAuthoritySet.Id:
		authoritiesProof := merkle.NewProof(authorityLeaves, authoritiesProof, uint64(cs.NextAuthoritySet.Len), hasher.Keccak256Hasher{})
		valid, err := authoritiesProof.Verify(cs.NextAuthoritySet.AuthorityRoot[:])
		if err != nil || !valid {
			return sdkerrors.Wrap(err, ErrAuthoritySetUnknown.Error())
		}
		updatedAuthority = true
		log.Printf("verify NextAuthoritySet successfully !\n")
	}

	log.Printf("cs.LatestBeefyHeight: %d\n", cs.LatestBeefyHeight)
	log.Printf("signedCommitment.Commitment.BlockNumer: %d\n", signedCommitment.Commitment.BlockNumer)
	// only update if we have a higher block number.
	if signedCommitment.Commitment.BlockNumer > cs.LatestBeefyHeight {
		for _, payload := range signedCommitment.Commitment.Payload {
			mmrRootID := []byte("mh")
			log.Printf("mmrRootID: %s\n", hex.EncodeToString(mmrRootID))
			log.Printf("payload.PayloadId: %s\n", hex.EncodeToString(payload.PayloadId[:]))
			// checks for the right payloadId
			if bytes.Equal(payload.PayloadId[:], mmrRootID) {
				// the next authorities are in the latest BeefyMmrLeaf

				// scale encode the mmr leaf
				mmrLeafBytes, err := scalecodec.Encode(mmrUpdateProof.MmrLeaf)
				if err != nil {
					return sdkerrors.Wrap(err, ErrInvalidCommitment.Error())
				}
				// we treat this leaf as the latest leaf in the mmr
				mmrSize := mmr.LeafIndexToMMRSize(mmrUpdateProof.MmrLeafIndex)
				mmrLeaves := []merkletypes.Leaf{
					{
						Hash:  crypto.Keccak256(mmrLeafBytes),
						Index: mmrUpdateProof.MmrLeafIndex,
					},
				}
				mmrProof := mmr.NewProof(mmrSize, mmrUpdateProof.MmrProof, mmrLeaves, hasher.Keccak256Hasher{})
				// verify that the leaf is valid, for the signed mmr-root-hash
				if !mmrProof.Verify(payload.PayloadData) {
					return sdkerrors.Wrap(err, ErrFailedVerifyMMRLeaf.Error()) // error!, mmr proof is invalid
				}
				log.Printf("verify relayer mmr root successfully !\n")
				// update the block_number
				cs.LatestBeefyHeight = signedCommitment.Commitment.BlockNumer
				log.Printf("updated cs.LatestBeefyHeight: %d\n", cs.LatestBeefyHeight)
				// updates the mmr_root_hash
				cs.MmrRootHash = payload.PayloadData

				log.Printf("updated cs.MmrRootHash : %s\n", hex.EncodeToString(cs.MmrRootHash))
				// authority set has changed, rotate our view of the authorities
				if updatedAuthority {
					cs.Authority = cs.NextAuthoritySet
					// mmr leaf has been verified, use it to update our view of the next authority set
					cs.NextAuthoritySet = &mmrUpdateProof.MmrLeaf.BeefyNextAuthoritySet
				}
				break
			}
		}
	}

	mmrProof, err := cs.parachainHeadersToMMRProof(beefyHeader)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to execute getMMRProf")
	}

	// Given the leaves, we should be able to verify that each parachain header was
	// indeed included in the leaves of our mmr.
	log.Printf("cs.MmrRootHash : %s\n", hex.EncodeToString(cs.MmrRootHash))
	root, err := mmrProof.CalculateRoot()
	if err != nil {
		log15.Error(fmt.Sprintf("failed to calculate root for mmr leaf %v", root))
		return sdkerrors.Wrap(err, ErrFailedEncodeMMRLeaf.Error())
	}

	log.Printf(" mmrProof.CalculateRoot() : %s !\n", hex.EncodeToString(root))
	if !mmrProof.Verify(cs.MmrRootHash) {
		log15.Error(fmt.Sprintf("failed to verify mmr leaf %v", root))
		return sdkerrors.Wrap(err, ErrFailedVerifyMMRLeaf.Error())
	}

	log.Printf("verify parachain header successfully !\n")
	return nil
}

func (cs *ClientState) parachainHeadersToMMRProof(beefyHeader *Header) (*mmr.Proof, error) {
	mmrLeaves := make([]merkletypes.Leaf, len(beefyHeader.ParachainHeaders))

	// verify parachain headers
	for i := 0; i < len(beefyHeader.ParachainHeaders); i++ {
		// first we need to reconstruct the mmr leaf for this header
		parachainHeader := beefyHeader.ParachainHeaders[i]

		headsLeafBytes, err := scalecodec.Encode(ParaIdAndHeader{ParaId: parachainHeader.ParaId, Header: parachainHeader.ParachainHeader})
		if err != nil {
			return nil, sdkerrors.Wrap(err, ErrInvalivParachainHeadsProof.Error())
		}
		headsLeaf := []merkletypes.Leaf{
			{
				Hash:  crypto.Keccak256(headsLeafBytes),
				Index: uint64(parachainHeader.HeadsLeafIndex),
			},
		}
		parachainHeadsProof := merkle.NewProof(headsLeaf, parachainHeader.ParachainHeadsProof, uint64(parachainHeader.HeadsTotalCount), hasher.Keccak256Hasher{})
		// todo: merkle.Proof.Root() should return fixed bytes
		parachainHeadsRoot, err := parachainHeadsProof.Root()
		// TODO: verify extrinsic root here once trie lib is fixed.
		if err != nil {
			return nil, sdkerrors.Wrap(err, ErrInvalivParachainHeadsProof.Error())
		}

		// not a fan of this but its golang
		var parachainHeads SizedByte32
		copy(parachainHeads[:], parachainHeadsRoot)

		mmrLeaf := BeefyMmrLeaf{
			Version:      parachainHeader.MmrLeafPartial.Version,
			ParentNumber: parachainHeader.MmrLeafPartial.ParentNumber,
			ParentHash:   parachainHeader.MmrLeafPartial.ParentHash,
			BeefyNextAuthoritySet: BeefyAuthoritySet{
				Id:            parachainHeader.MmrLeafPartial.BeefyNextAuthoritySet.Id,
				AuthorityRoot: parachainHeader.MmrLeafPartial.BeefyNextAuthoritySet.AuthorityRoot,
				Len:           parachainHeader.MmrLeafPartial.BeefyNextAuthoritySet.Len,
			},
			ParachainHeads: &parachainHeads,
		}

		// the mmr leaf's are a scale-encoded
		mmrLeafBytes, err := scalecodec.Encode(mmrLeaf)
		if err != nil {
			return nil, sdkerrors.Wrap(err, ErrInvalidMMRLeaf.Error())
		}

		mmrLeaves[i] = merkletypes.Leaf{
			Hash: crypto.Keccak256(mmrLeafBytes),
			// based on our knowledge of the beefy protocol, and the structure of MMRs
			// we are be able to reconstruct the leaf index of this mmr leaf
			// given the parent_number of this leaf, the beefy activation block
			Index: uint64(cs.GetLeafIndexForBlockNumber(parachainHeader.MmrLeafPartial.ParentNumber + 1)),
		}
	}

	mmrProof := mmr.NewProof(beefyHeader.MmrSize, beefyHeader.MmrProofs, mmrLeaves, hasher.Keccak256Hasher{})

	return mmrProof, nil
}

func (cs ClientState) UpdateState(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, clientMsg exported.ClientMessage) []exported.Height {
	header, _ := clientMsg.(*Header)
	height := header.GetHeight().(clienttypes.Height)
	return []exported.Height{height}
}

// pruneOldestConsensusState will retrieve the earliest consensus state for this clientID and check if it is expired. If it is,
// that consensus state will be pruned from store along with all associated metadata. This will prevent the client store from
// becoming bloated with expired consensus states that can no longer be used for updates and packet verification.
func (cs ClientState) pruneOldestConsensusState(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore) {
	// Check the earliest consensus state to see if it is expired, if so then set the prune height
	// so that we can delete consensus state and all associated metadata.
	// var (
	// 	pruneHeight exported.Height
	// )

	// pruneCb := func(height exported.Height) bool {
	// 	consState, found := GetConsensusState(clientStore, cdc, height)
	// 	// this error should never occur
	// 	if !found {
	// 		panic(sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "failed to retrieve consensus state at height: %s", height))
	// 	}

	// 	if cs.IsExpired(consState.Timestamp, ctx.BlockTime()) {
	// 		pruneHeight = height
	// 	}

	// 	return true
	// }

	// IterateConsensusStateAscending(clientStore, pruneCb)

	// // if pruneHeight is set, delete consensus state and metadata
	// if pruneHeight != nil {
	// 	deleteConsensusState(clientStore, pruneHeight)
	// 	deleteConsensusMetadata(clientStore, pruneHeight)
	// }
	panic("implement me")
}

// UpdateStateOnMisbehaviour updates state upon misbehaviour, freezing the ClientState. This method should only be called when misbehaviour is detected
// as it does not perform any misbehaviour checks.
func (cs ClientState) UpdateStateOnMisbehaviour(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, _ exported.ClientMessage) {
	clientStore.Set(host.ClientStateKey(), clienttypes.MustMarshalClientState(cdc, &cs))

	panic("implement me")
}
