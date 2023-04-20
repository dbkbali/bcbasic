package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyPairSignVerifyValid(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	msg := []byte("Hello World")
	sig, err := privKey.Sign(msg)
	assert.Nil(t, err)

	assert.True(t, sig.Verify(pubKey, msg))
}

func TestKeyPairSignVerifyFail(t *testing.T) {
	privKey, _ := GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	msg := []byte("Hello World")
	sig, err := privKey.Sign(msg)
	assert.Nil(t, err)

	otherPrivKey, _ := GeneratePrivateKey()
	otherPubKey := otherPrivKey.PublicKey()

	assert.False(t, sig.Verify(otherPubKey, msg))
	assert.False(t, sig.Verify(pubKey, []byte("Hello World!")))
}
