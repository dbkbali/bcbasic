package types

import "encoding/hex"

type Address [20]uint8

func (a Address) ToSlice() []byte {
	b := make([]byte, 20)
	for i := 0; i < 20; i++ {
		b[i] = a[i]
	}

	return b
}

func (a Address) String() string {
	return hex.EncodeToString(a.ToSlice())
}

func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		panic("AddressFromBytes: invalid length")
	}

	var value [20]uint8
	for i := 0; i < 20; i++ {
		value[i] = b[i]
	}

	return Address(value)
}
