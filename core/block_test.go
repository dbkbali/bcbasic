package core

import (
	"bytes"
	"testing"
	"time"

	"github.com/dbkbali/bcbasic/crypto"
	"github.com/dbkbali/bcbasic/types"
	"github.com/stretchr/testify/assert"
)

func TestSignBLock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)

}

func TestVerifyBLock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())

	b.Height = 100
	assert.NotNil(t, b.Verify())
}

func TestBlockEncodeDecode(t *testing.T) {
	b := randomBlock(t, 1, types.Hash{})
	buf := &bytes.Buffer{}
	assert.Nil(t, b.Encode(NewGobBlockEncoder(buf)))

	bDecode := new(Block)
	assert.Nil(t, bDecode.Decode(NewGobBlockDecoder(buf)))
	assert.Equal(t, b, bDecode)

}

func randomBlock(t *testing.T, height uint32, prevBlockhash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	header := &Header{
		Version:       1,
		PrevBlockHash: prevBlockhash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}

	b, err := NewBlock(header, []*Transaction{tx})
	assert.Nil(t, err)
	dataHash, err := CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash
	assert.Nil(t, b.Sign(privKey))

	return b
}
