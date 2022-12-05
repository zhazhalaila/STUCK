package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/stuck/crypto/tbls/key"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
)

func writeToFile(filePrefix string, n int, fileName string, shareBytes []byte) {
	err := ioutil.WriteFile("../keys/"+filePrefix+"/"+strconv.Itoa(n)+fileName, shareBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	n := flag.Int("n", 4, "total node number")
	t := flag.Int("t", 2, "threshold of tbls")

	flag.Parse()
	f := (*n - 1) / 3

	suite := bn256.NewSuite()
	priShares := make([]key.PriShare, *n) // Each node has its own private key
	pubShares := make([]key.PubShare, *n) // A public key will be split into n shares

	secret := suite.G1().Scalar().Pick(suite.RandomStream())
	priPoly := share.NewPriPoly(suite.G2(), *t, secret, suite.RandomStream()) // Private key set
	pubPoly := priPoly.Commit(suite.G2().Point().Base())                      // Public key

	// Marshal private key
	for i, x := range priPoly.Shares(*n) {
		priBytes, err := x.V.MarshalBinary()
		if err != nil {
			log.Fatal(err)
		}
		ps := key.PriShare{Index: x.I, PriBytes: priBytes}
		priShares[i] = ps
	}

	// Marshal private key set
	priSharesBytes, err := json.Marshal(priShares)
	if err != nil {
		log.Fatal(err)
	}

	// Write private key set to file
	var filePrefix string
	if *t == f+1 {
		filePrefix = "f+1"
	} else {
		filePrefix = "2f+1"
	}
	writeToFile(filePrefix, *n, "/private_key.conf", priSharesBytes)

	// Marshal public key
	for i, x := range pubPoly.Shares(*n) {
		pubShareBytes, err := x.V.MarshalBinary()
		if err != nil {
			log.Fatal(err)
		}
		ps := key.PubShare{Index: x.I, PubBytes: pubShareBytes}
		pubShares[i] = ps
	}

	// Marshal public key shares
	pubSharesBytes, err := json.Marshal(pubShares)
	if err != nil {
		log.Fatal(err)
	}

	// Write public key set to file
	writeToFile(filePrefix, *n, "/public_key.conf", pubSharesBytes)
}
