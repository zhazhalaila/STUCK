package decode

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/stuck/crypto/tbls/key"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
)

func getFilePrefix(n, t int) string {
	f := (n - 1) / 3
	if t == f+1 {
		return "f+1"
	} else {
		return "2f+1"
	}
}

func readFromFile(filePrefix string, n int, fileName string) ([]byte, error) {
	plan, err := ioutil.ReadFile("../crypto/tbls/keys/" + filePrefix + "/" + strconv.Itoa(n) + fileName)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// Decode private share for a specific node
func DecodePriShare(n, t, i int) *share.PriShare {
	filePrefix := getFilePrefix(n, t)
	plan, err := readFromFile(filePrefix, n, "/private_key.conf")
	if err != nil {
		log.Fatal(err)
	}

	var priShares []key.PriShare
	err = json.Unmarshal(plan, &priShares)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal
	suite := bn256.NewSuite()
	scalar := suite.G2().Scalar()
	err = scalar.UnmarshalBinary(priShares[i].PriBytes)
	if err != nil {
		log.Fatal(err)
	}

	return &share.PriShare{I: priShares[i].Index, V: scalar}
}

// Decode public key
func DecodePubShare(n, t int) *share.PubPoly {
	filePrefix := getFilePrefix(n, t)
	plan, err := readFromFile(filePrefix, n, "//public_key.conf")
	if err != nil {
		log.Fatal(err)
	}

	var pubSharesBytes []key.PubShare
	err = json.Unmarshal(plan, &pubSharesBytes)
	if err != nil {
		log.Fatal(err)
	}

	pubShares := make([]*share.PubShare, n)
	suite := bn256.NewSuite()

	// Unmarshal public shares
	for i, b := range pubSharesBytes {
		point := suite.G2().Point()
		pubShares[i] = &share.PubShare{}
		pubShares[i].I = b.Index
		err = point.UnmarshalBinary(b.PubBytes)
		if err != nil {
			log.Fatal(err)
		}
		pubShares[i].V = point
	}

	// Recovery public key
	pubPoly, err := share.RecoverPubPoly(suite.G2(), pubShares, t, n)
	if err != nil {
		log.Fatal(err)
	}

	return pubPoly
}
