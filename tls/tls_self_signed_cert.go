package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"time"
)

func main() {
	n := flag.Int("n", 1, "total node number")
	flag.Parse()

	for i := 1; i <= *n; i++ {
		fmt.Println(i)
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			log.Fatalf("Failed to generate serial number: %v", err)
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"order.org"},
			},
			DNSNames:  []string{"order" + strconv.Itoa(i) + ".com"},
			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(24 * 30 * 12 * time.Hour),

			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			BasicConstraintsValid: true,
		}

		// Create self-signed certificate.
		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			log.Fatalf("Failed to create certificate: %v", err)
		}

		pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		if pemCert == nil {
			log.Fatal("Failed to encode certificate to PEM")
		}

		if err = ioutil.WriteFile("../certs/cert"+strconv.Itoa(i)+".pem", pemCert, 0644); err != nil {
			log.Fatal(err)
		}

		log.Printf("wrote cert.pem for [%d] node.\n", i)

		privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			log.Fatalf("Unable to marshal private key: %v", err)
		}
		pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
		if pemKey == nil {
			log.Fatal("Failed to encode key to PEM")
		}

		if err = ioutil.WriteFile("../certs/key"+strconv.Itoa(i)+".pem", pemKey, 0644); err != nil {
			log.Fatal(err)
		}

		log.Printf("wrote key.pem for [%d] node.\n", i)
	}
}
