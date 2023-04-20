package core

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/dbkbali/bcbasic/types"
	"github.com/stretchr/testify/assert"
)

func TestHeader_Encode_Decode(t *testing.T) {
	h := Header{
		Version:   1,
		PrevBlock: types.RandomHash(),
		Timestamp: time.Now().UnixNano(),
		Height:    10,
		Nonce:     9980908909,
	}

	buf := &bytes.Buffer{}
	assert.Nil(t, h.EncodeBinary(buf))

	hDecode := Header{}
	assert.Nil(t, hDecode.DecodeBinary(buf))
	assert.Equal(t, h, hDecode)
	fmt.Printf("%+v", hDecode)
}

func TestBlock_Encode_Decode(t *testing.T) {
	b := Block{
		Header: Header{
			Version:   1,
			PrevBlock: types.RandomHash(),
			Timestamp: time.Now().UnixNano(),
			Height:    10,
			Nonce:     9980908909,
		},
		Transactions: []Transaction{
			{},
			{},
			{},
		},
	}

	buf := &bytes.Buffer{}
	assert.Nil(t, b.EncodeBinary(buf))

	bDecode := Block{}
	assert.Nil(t, bDecode.DecodeBinary(buf))
	assert.Equal(t, b, bDecode)
	fmt.Printf("%+v", bDecode)
}

func TestBlock_Hash(t *testing.T) {
	b := Block{
		Header: Header{
			Version:   1,
			PrevBlock: types.RandomHash(),
			Timestamp: time.Now().UnixNano(),
			Height:    10,
			Nonce:     9980908909,
		},
		Transactions: []Transaction{
			{},
			{},
			{},
		},
	}

	h := b.Hash()
	assert.False(t, h.IsZero())
}
