package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

/**
References:

https://github.com/pquerna/otp/

**/

type Configuration struct {
	AppName     string `json:"appName,omitempty"`
	AccountName string `json:"accountName,omitempty"`
	SecretKey   string `json:"secretKey"`
}

func loadConfig(cPtr string) Configuration {

	var c Configuration
	fmt.Println("Loading the following config file: " + cPtr + "\n")
	// go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(cPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	// var config Configuration
	if err := decoder.Decode(&c); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	return c
}

func readStringFromPrompt(prompt string) string {
	//fmt.Printf("Input OTP to Validate: ")
	var message string
	fmt.Printf("%s", prompt)
	if _, err := fmt.Scan(&message); err != nil {
		cf.CheckError("Unable to read input", err, true)
	}
	return message
}

type ValidateOptions struct {
	Code          string // Gathered from user input from the google authenticator
	SecretKey     string // Read from the config.json
	CurrentTime   time.Time
	HashValidTime int    // 30 seconds is the default
	TimeSkew      int    // 0 is the Default - Could be 1 for some tolerance
	CodeLength    int    // 6 is Default - Length of the code Input
	Algorithm     string // SHA1 is default
}

func setValidateOptions(otpInput string, secretKey string) ValidateOptions {
	var v ValidateOptions
	v.Code = otpInput
	v.SecretKey = secretKey
	//v.CurrentTime = time.Now().UTC().Add((-30 * time.Second) * 11) // I am not sure why the time shift exists that is about 5 minutes and 30 seconds
	v.CurrentTime = time.Now().UTC() // The time shift was somehow my laptop was not automatically setting the time and it had drifted over time about 5 minutes and 30 seconds
	v.HashValidTime = 30
	v.TimeSkew = 0 // It can be set to 1 for tolerance - Tolerance is 30 seconds back and 30 seconds forward...
	v.CodeLength = 6
	v.Algorithm = "SHA1"

	return v
}

func containsOnlyNumbers(input string) bool {
	for _, char := range input {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func validateCode(v ValidateOptions) (bool, error) {
	counters := []uint64{}
	counter := int64(math.Floor(float64(v.CurrentTime.Unix()) / float64(v.HashValidTime)))

	counters = append(counters, uint64(counter))
	for i := 1; i <= int(v.TimeSkew); i++ {
		counters = append(counters, uint64(counter+int64(i)))
		counters = append(counters, uint64(counter-int64(i)))
	}

	// Santize the code - Remove the line breaks and spaces
	sCode := strings.ReplaceAll(v.Code, "\n", "")
	sCode = strings.ReplaceAll(sCode, "\r", "")
	sCode = strings.ReplaceAll(sCode, " ", "")
	//fmt.Printf("\nDebug Input Code: %s\n", sCode)

	// Verify the length of the code
	if len(sCode) != v.CodeLength {
		return false, errors.New("code length does not equal 6")
	}

	// Verify the code is all digits
	if !containsOnlyNumbers(sCode) {
		return false, errors.New("code needs to contain only numbers")
	}

	validCode := false
	for _, counter := range counters {
		generatedPasscode, err := generateCodeFromSecret(v, counter)
		if err != nil {
			return false, errors.New("unable to generate passcode from secret")
		}
		// The generated code truncates leading 0's
		if len(generatedPasscode) < 6 {
			if len(generatedPasscode) == 3 {
				generatedPasscode = "000" + generatedPasscode
			} else if len(generatedPasscode) == 4 {
				generatedPasscode = "00" + generatedPasscode
			} else if len(generatedPasscode) == 5 {
				generatedPasscode = "0" + generatedPasscode
			}
		}
		//fmt.Printf("Debug Generated Code: %s\n", generatedPasscode)

		if sCode == generatedPasscode {
			validCode = true
		}
	}

	if validCode {
		return true, nil
	}

	return false, errors.New("code is invalid")
}

func generateCodeFromSecret(v ValidateOptions, counter uint64) (passcode string, err error) {
	secret := strings.TrimSpace(v.SecretKey)
	if n := len(secret) % 8; n != 0 {
		secret = secret + strings.Repeat("=", 8-n)
	}
	//fmt.Printf("\nDebug SecretKey: %s\n", secret)

	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", errors.New("unable to decode the secret to bytes")
	}

	buf := make([]byte, 8)
	mac := hmac.New(sha1.New, secretBytes)
	binary.BigEndian.PutUint64(buf, counter)

	mac.Write(buf)
	sum := mac.Sum(nil)

	// "Dynamic truncation" in RFC 4226
	// http://tools.ietf.org/html/rfc4226#section-5.4
	offset := sum[len(sum)-1] & 0xf
	value := int64(((int(sum[offset]) & 0x7f) << 24) |
		((int(sum[offset+1] & 0xff)) << 16) |
		((int(sum[offset+2] & 0xff)) << 8) |
		(int(sum[offset+3]) & 0xff))
	//fmt.Printf("Debug Value: %d\n", value)

	passcodeInt := int32(value % int64(math.Pow10(6)))
	// The below could be wrong...
	//return fmt.Sprintf("%%0%dd", passcodeInt), nil
	return fmt.Sprintf("%d", passcodeInt), nil

}

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	config = loadConfig(*ConfigPtr)
	otpInput := readStringFromPrompt("Input OTP to Validate: ")
	validateOptions := setValidateOptions(otpInput, config.SecretKey)
	valid, err := validateCode(validateOptions)
	cf.CheckError("Validating the input code had an issue", err, true)

	if valid {
		fmt.Println("Code is Validated")
	} else {
		fmt.Println("ERROR: Code is invalid...")
	}
}
