package crypto

// Cryptographic Interface
type CryptoProvider interface {
	ComputeShare(data []byte) []byte                    // compute partial share
	VerifyShare(data []byte, share []byte) bool         // verify partial share
	Combine(data []byte, shares [][]byte) []byte        // combine signature
	VerifySignature(data []byte, signature []byte) bool // verify signature
}
