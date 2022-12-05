package identity

import (
	"github.com/stuck/crypto/tbls/decode"

	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

type TBLSCryproProvider struct {
	n       int
	t       int
	suite   *bn256.Suite
	priPoly *share.PriShare
	pubPoly *share.PubPoly
}

func NewTBLSCryproProvider(n, t, i int) *TBLSCryproProvider {
	cp := &TBLSCryproProvider{}
	cp.n = n
	cp.t = t
	cp.suite = bn256.NewSuite()
	cp.priPoly = decode.DecodePriShare(n, t, i)
	cp.pubPoly = decode.DecodePubShare(n, t)
	return cp
}

// Compute partial signature
func (cp *TBLSCryproProvider) ComputeShare(data []byte) []byte {
	share, err := tbls.Sign(cp.suite, cp.priPoly, data)
	if err != nil {
		return nil
	}
	return share
}

// Verify partial signature
func (cp *TBLSCryproProvider) VerifyShare(data []byte, share []byte) bool {
	err := tbls.Verify(cp.suite, cp.pubPoly, data, share)
	return err == nil
}

// Combine signature
func (cp *TBLSCryproProvider) Combine(data []byte, shares [][]byte) []byte {
	signature, err := tbls.Recover(cp.suite, cp.pubPoly, data, shares, cp.t, cp.n)
	if err != nil {
		return nil
	}
	return signature
}

// Verify signature
func (cp *TBLSCryproProvider) VerifySignature(data []byte, signature []byte) bool {
	err := bls.Verify(cp.suite, cp.pubPoly.Commit(), data, signature)
	return err == nil
}
