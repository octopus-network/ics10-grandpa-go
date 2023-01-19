package grandpa_test

import (
	"log"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"

	"github.com/stretchr/testify/require"
)

func TestBeefyAuthoritySetCodec(t *testing.T) {
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
