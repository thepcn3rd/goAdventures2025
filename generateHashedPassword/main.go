package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

// GenerateRandomString generates a random string of a given length
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	var result string

	for i := 0; i < length; i++ {
		// Generate a random index within the range of the charset
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}
		result += string(charset[randomIndex.Int64()])
	}

	return result, nil
}

func GenerateSalt(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	// Encode to a URL-safe base64 string and trim to the required length
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

func sha512Hash(password string) string {
	hasher := sha512.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

func sha512HashCrypt(password string) string {
	salt, _ := GenerateSalt(16)
	cmd := exec.Command("openssl", "passwd", "-6", "-salt", salt, password)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

type PrivateKey interface {
	Public() crypto.PublicKey
	Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
}

func generateRSAKey() (PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func generateECDSA256Key() (PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func generateECDSA384Key() (PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
}

func sshPublicKey(keyType string) string {
	// Generate a new RSA private key
	var privateKey PrivateKey
	var err error
	switch keyType {
	case "rsa":
		privateKey, err = generateRSAKey()
	case "ecdsa256":
		privateKey, err = generateECDSA256Key()
	case "ecdsa384":
		privateKey, err = generateECDSA384Key()
	default:
		fmt.Println("Invalid choice")
		return ""
	}
	if err != nil {
		log.Fatalln("Error generating private key:", err)
	}

	// Generate the public key in OpenSSH format
	publicKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		log.Fatalln("Error creating OpenSSH public key:", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	return string(publicKeyBytes)
}

func main() {
	passPTR := flag.String("p", "", "Input a password to hash")
	flag.Parse()

	password := *passPTR
	password = strings.Replace(password, "\n", "", -1)
	password = strings.Replace(password, "\r", "", -1)

	for i := 0; i <= 10; i++ {
		var err error
		//fmt.Println("Password: ", password)
		if len(password) == 0 {
			password, err = GenerateRandomString(32)
			if err != nil {
				fmt.Println("Error generating random password:", err)
				return
			}
		}

		//fmt.Printf("Input/Generated Password: %s\n", password)

		// Hash the generated password
		//hashedPassword := sha512Hash(password)
		//hashedPassword := sha512HashCrypt(password)
		//hashedPassword := sshPublicKey("rsa")
		//hashedPassword := sshPublicKey("ecdsa256")
		hashedPassword := sshPublicKey("ecdsa384")
		//fmt.Printf("Hashed Password: %s\n", hashedPassword)
		fmt.Printf("%s\n", hashedPassword)
	}
}
