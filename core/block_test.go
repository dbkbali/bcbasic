package core

import (
	"testing"
	"time"

	"github.com/dbkbali/bcbasic/crypto"
	"github.com/dbkbali/bcbasic/types"
	"github.com/stretchr/testify/assert"
)

func RandomBlock(height uint32) *Block {
	header := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}
	tx := Transaction{
		Data: []byte("foobar"),
	}

	return NewBlock(header, []Transaction{tx})
}

func TestSignBLock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	b := RandomBlock(0)

	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)

}

func TestVerifyBLock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	b := RandomBlock(0)

	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())

	b.Height = 100
	assert.NotNil(t, b.Verify())
}
