package core

import (
	"bytes"
	"testing"

	"github.com/dbkbali/bcbasic/crypto"
	"github.com/stretchr/testify/assert"
)

func TestSignTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	data := []byte("foobar")
	tx := &Transaction{
		Data: data,
	}

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)

	// TODO
}

func TestVerifyTransaction(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()

	data := []byte("foobar")
	tx := &Transaction{
		Data: data,
	}

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)

	assert.Nil(t, tx.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()

	assert.NotNil(t, tx.Verify())
}

func TestTxEncodeDecode(t *testing.T) {
	tx := randomTxWithSignature(t)

	buf := &bytes.Buffer{}

	enc := NewGobTxEncoder(buf)
	assert.Nil(t, enc.Encode(tx))

	decTx := new(Transaction)
	assert.Nil(t, decTx.Decode(NewGobTxDecoder(buf)))
	assert.Equal(t, tx, decTx)
}

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("foobar"),
	}

	assert.Nil(t, tx.Sign(privKey))

	return tx
}
