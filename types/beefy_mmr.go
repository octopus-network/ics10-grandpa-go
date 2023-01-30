package grandpa

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type VersionedFinalityProof struct {
	Version          uint8
	SignedCommitment types.SignedCommitment
}

type ParaIdAndHeader struct {
	ParaId uint32
	Header []byte
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

type SizedByte32 [32]byte

func (b *SizedByte32) Marshal() ([]byte, error) {
	return b[:], nil
}

func (b *SizedByte32) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	copy(b[:], data)
	return nil
}

func (b *SizedByte32) Size() int {
	return len(b)
}

type SizedByte2 [2]byte

func (b *SizedByte2) Marshal() ([]byte, error) {
	return b[:], nil
}

func (b *SizedByte2) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	copy(b[:], data)
	return nil
}

func (b *SizedByte2) Size() int {
	return len(b)
}

type U8 uint8

func (u *U8) Marshal() ([]byte, error) {
	return []byte{byte(*u)}, nil
}

func (u *U8) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	*u = U8(data[0])
	return nil
}

func (u *U8) Size() int {
	return 1
}

func Bytes32(bytes []byte) SizedByte32 {
	var buffer SizedByte32
	copy(buffer[:], bytes)
	return buffer
}
