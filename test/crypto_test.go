package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stuck/crypto"
	"github.com/stuck/crypto/tbls/identity"
)

// Test if TBLS is working
func TestTBLSWith4Nodes(t *testing.T) {
	players := make([]crypto.CryptoProvider, 4)
	for i := range players {
		players[i] = identity.NewTBLSCryproProvider(4, 2, i)
	}

	// Byzantine node tampering share
	msg := []byte("Hello world")
	byzantine := players[0]
	share0 := byzantine.ComputeShare(msg)
	assert.Equal(t, true, players[1].VerifyShare(msg, share0))
	share0[0] ^= 0x10
	assert.Equal(t, false, players[1].VerifyShare(msg, share0))

	// Two honest nodes can generate valid signature
	honest1 := players[1]
	honest2 := players[2]
	share1 := honest1.ComputeShare(msg)
	share2 := honest2.ComputeShare(msg)
	shares := make([][]byte, 2)
	shares[0] = share1
	shares[1] = share2
	signature := players[3].Combine(msg, shares)
	assert.NotNil(t, signature)
	assert.Equal(t, true, players[3].VerifySignature(msg, signature))
}
