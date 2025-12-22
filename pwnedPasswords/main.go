package main

import (
	"bufio"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
	"golang.org/x/crypto/md4"
	"golang.org/x/text/encoding/unicode"
)

/**

This program evaluates SHA1 and NTLM Hashes using the k-Anonimity model to see if they
exist in the haveibeenpwned online API

(Done) 1. Lookup input of password (not at the command line so that it is in the history)
(Done) 2. Lookup input of SHA1 or NTLM Hash
(Done) 3. Look in offline file and then pull, then update offline file
(Done) 4. Read hashes from a file
5. Add custom hashes to offline file based on wordlist or permutations...
(Done) 6. Read plain-text passwords from a file distinguished by -sha1 or -ntlm and then the filename

Found that the logic of the comparison of the hashes should be done in upper-case.  Was missing some matches.  Went through and verified in an if statement
the prefix and the suffix are made to be upper-case...

Found that duplicate prefixes can exist if not identified as duplicates and combined.  This
occurs when the suffix does not exist in the initial offline file scan and then is found in a new
pull of the API.  This adds a second entry in the offline struct.  testCase01 - Tested with the
Welcome123 password, removed the specific hash and then ran the program.  It recognized the prefix was
a duplicate and combined them and removed the duplicates.

**/

type Configuration struct {
	URL                  string `json:"url"`
	RequestsDelay        int    `json:"requestsDelay"`
	UserAgent            string `json:"userAgent"`
	OfflineFiles         string `json:"offlineFilesDirectory"`
	SkipLoadOfflineFiles bool   `json:"skipLoadOfflineFiles"`
	SkipSaveOfflineFiles bool   `json:"skipSaveOfflineFiles"`
}

// Based the offline lookup on the k-Anonymity Model
type HashOfflineLookupStruct struct {
	SHA1HashPrefix []PrefixStruct `json:"sha1Prefix"`
	NTLMHashPrefix []PrefixStruct `json:"ntlmPrefix"`
}

type PrefixStruct struct {
	Prefix      string   `json:"prefix"`
	Suffix      []string `json:"suffix"`
	Created     string   `json:"created"`
	LastUpdated string   `json:"lastUpdated"`
}

func (c *Configuration) CreateConfig() error {
	c.URL = "https://api.pwnedpasswords.com/range/"
	c.UserAgent = "goPwnedPasswords-Check-v1"
	c.RequestsDelay = 3
	c.OfflineFiles = "offlineFiles"
	c.SkipLoadOfflineFiles = false
	c.SkipSaveOfflineFiles = false

	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile("config.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Configuration) LoadConfig(cPtr string) error {
	configFile, err := os.Open(cPtr)
	if err != nil {
		return err
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&c); err != nil {
		return err
	}

	return nil
}

func (h *HashOfflineLookupStruct) CreateOfflineFiles(c Configuration) error {

	if len(h.SHA1HashPrefix) > 0 {
		for _, sha1 := range h.SHA1HashPrefix {
			jsonData, err := json.MarshalIndent(sha1, "", "    ")
			if err != nil {
				return err
			}
			saveFileName := c.OfflineFiles + "/sha1/offline_" + sha1.Prefix + ".json"
			err = os.WriteFile(saveFileName, jsonData, 0644)
			if err != nil {
				return err
			}
		}
	}

	if len(h.NTLMHashPrefix) > 0 {
		for _, ntlm := range h.NTLMHashPrefix {
			jsonData, err := json.MarshalIndent(ntlm, "", "    ")
			if err != nil {
				return err
			}
			saveFileName := c.OfflineFiles + "/ntlm/offline_" + ntlm.Prefix + ".json"
			err = os.WriteFile(saveFileName, jsonData, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RemoveBadChars(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}

// HashPassword hashes the password using SHA-1
func SHA1Hash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}

func NTLMHash(password string) string {
	// Convert the password to UTF-16 little-endian
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	encoder := utf16.NewEncoder()
	passwordBytes, err := encoder.Bytes([]byte(password))
	if err != nil {
		log.Fatalf("Failed to encode password: %v", err)
	}

	// Compute the MD4 hash of the UTF-16 password
	hash := md4.New()
	hash.Write(passwordBytes)
	hashSum := hash.Sum(nil)

	// Return the hash as a hexadecimal string
	return strings.ToUpper(hex.EncodeToString(hashSum))
}

func InputFromStdin() string {

	fmt.Println("Interactive Mode - Select Option")
	fmt.Println("1. Input plain text password")
	fmt.Println("2. Input SHA1 or NTLM hash")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("[E] Error reading input:", err)
	}
	input = RemoveBadChars(input)
	switch input {
	case "1":
		fmt.Println("\nSelect Hash to use for the Password")
		fmt.Println("1. SHA1")
		fmt.Println("2. NTLM")
		readerType := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		inputType, err := readerType.ReadString('\n')
		if err != nil {
			log.Fatalln("[E] Error reading input for selection of hash type:", err)
		}
		inputType = RemoveBadChars(inputType)

		readerPassword := bufio.NewReader(os.Stdin)
		fmt.Println("\nInput Plain-text Password")
		fmt.Print("> ")
		inputPassword, err := readerPassword.ReadString('\n')
		if err != nil {
			log.Fatalln("[E] Error reading input of the plain text password:", err)
		}
		fmt.Println()
		inputPassword = RemoveBadChars(inputPassword)

		if inputType == "1" {
			return SHA1Hash(inputPassword)
		} else if inputType == "2" {
			return NTLMHash(inputPassword)
		}
	case "2":
		fmt.Println("\nInput Hash")
		readerType := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		inputType, err := readerType.ReadString('\n')
		if err != nil {
			log.Fatalln("[E] Error reading input for selection of hash:", err)
		}
		inputType = RemoveBadChars(inputType)
		fmt.Println()
		if len(inputType) == 40 || len(inputType) == 32 {
			return inputType
		} else {
			log.Fatalln("[E] Error reading the hash, length appears to be incorrect", err)
		}
	}

	log.Fatalln("[E] Error reading the input of a password or hash...", err)
	return ""
}

// Extending the function to support input of hashes, sha1passwords or ntlmpasswords
func InputFromFile(f string, fileType string) []string {
	outputStrings := []string{}
	fmt.Printf("\n[*] Processing File: %s\n\n", f)
	file, err := os.Open(f)
	if err != nil {
		log.Printf("[E] Failed to open file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Could introduce parsing of the line here...
		line = RemoveBadChars(line)
		if fileType == "hashes" && (len(line) == 40 || len(line) == 32) {
			outputStrings = append(outputStrings, line)
		} else if fileType == "sha1passwords" {
			outputStrings = append(outputStrings, SHA1Hash(line))
		} else if fileType == "ntlmpasswords" {
			outputStrings = append(outputStrings, NTLMHash(line))
		}
	}
	return outputStrings
}

func removeDuplicateSuffix(suffix []string) []string {

	seen := make(map[string]bool)
	result := []string{}

	for _, str := range suffix {
		if _, ok := seen[str]; !ok {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

// CheckPassword checks if the password has been pwned using HIBP API
func CheckHash(hashInput string, c Configuration) (bool, error, bool) {
	// Skip the verification of a self-signed certificate
	// Only connect if TLS1.2 or TLS1.3 is negotiated with a provided cipher
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
		},
	}
	client := &http.Client{Transport: tr}
	var url string
	var hashType string
	foundHash := false
	foundHashOffline := false
	//log.Println(hashInput)
	hashPrefix := hashInput[:5]
	suffix := hashInput[5:]

	if len(hashInput) == 40 {
		url = fmt.Sprintf("%s%s", c.URL, hashPrefix)
		hashType = "SHA1"
	} else {
		url = fmt.Sprintf("%s%s?mode=ntlm", c.URL, hashPrefix)
		hashType = "NTLM"
	}

	// Check Offline Database
	if !c.SkipLoadOfflineFiles && hashType == "SHA1" {
		for _, prefix := range HashStruct.SHA1HashPrefix {
			if strings.ToUpper(prefix.Prefix) == strings.ToUpper(hashPrefix) && !foundHash {
				for _, suffixOffline := range prefix.Suffix {
					if strings.ToUpper(suffixOffline) == strings.ToUpper(suffix) && !foundHash {
						foundHash = true
					}
				}
			}
		}
	}

	if !c.SkipLoadOfflineFiles && hashType == "NTLM" {
		for _, prefix := range HashStruct.NTLMHashPrefix {
			if strings.ToUpper(prefix.Prefix) == strings.ToUpper(hashPrefix) && !foundHash {
				for _, suffixOffline := range prefix.Suffix {
					if strings.ToUpper(suffixOffline) == strings.ToUpper(suffix) && !foundHash {
						foundHash = true
					}
				}
			}
		}
	}

	// Verify that hashes are found in the offline file struct
	if foundHash {
		//	fmt.Println("[*] Found hash in an offline file...")
		foundHashOffline = true
	}

	// Check haveibeenpwned API if the hash is not found in offlineFiles
	if !foundHash {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, err, false
		}
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			return false, err, false
		}
		defer resp.Body.Close()
		// Put in the delay
		time.Sleep(time.Duration(c.RequestsDelay) * time.Second)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err, false
		}
		//fmt.Printf("%s\n\n", string(body))
		hashes := strings.Split(string(body), "\n")
		var prefix PrefixStruct
		prefix.Prefix = hashPrefix
		currentTime := time.Now()
		formattedTime := currentTime.Format("2006-01-02 15:04:05")
		prefix.Created = formattedTime

		for _, h := range hashes {
			parts := strings.Split(h, ":")
			prefix.Suffix = append(prefix.Suffix, strings.ToUpper(parts[0]))
			if len(parts) < 2 {
				continue
			}
			if strings.ToUpper(parts[0]) == suffix {
				//return true, nil
				foundHash = true
			}
		}
		// Populate the offlineStruct
		if hashType == "SHA1" {
			prefixExists := false
			for i, h := range HashStruct.SHA1HashPrefix {
				if h.Prefix == prefix.Prefix {
					fmt.Println("[*] Prefix Exists in Struct Combining and Removing Duplicates")
					prefixExists = true
					HashStruct.SHA1HashPrefix[i].Suffix = append(HashStruct.SHA1HashPrefix[i].Suffix, prefix.Suffix...)
					HashStruct.SHA1HashPrefix[i].Suffix = removeDuplicateSuffix(HashStruct.SHA1HashPrefix[i].Suffix)
					currentTime := time.Now()
					// Update the Last Updated Time Stamp in the Struct
					formattedTime := currentTime.Format("2006-01-02 15:04:05")
					HashStruct.SHA1HashPrefix[i].LastUpdated = formattedTime
				}
			}
			if !prefixExists {
				HashStruct.SHA1HashPrefix = append(HashStruct.SHA1HashPrefix, prefix)
			}
			//HashStruct.SHA1HashPrefix = append(HashStruct.SHA1HashPrefix, prefix)
		} else {
			prefixExists := false
			for i, h := range HashStruct.NTLMHashPrefix {
				if h.Prefix == prefix.Prefix {
					fmt.Println("[*] Prefix Exists in Struct Combining and Removing Duplicates")
					prefixExists = true
					HashStruct.NTLMHashPrefix[i].Suffix = append(HashStruct.NTLMHashPrefix[i].Suffix, prefix.Suffix...)
					HashStruct.NTLMHashPrefix[i].Suffix = removeDuplicateSuffix(HashStruct.NTLMHashPrefix[i].Suffix)
					currentTime := time.Now()
					// Update the Last Updated Time Stamp in the Struct
					formattedTime := currentTime.Format("2006-01-02 15:04:05")
					HashStruct.NTLMHashPrefix[i].LastUpdated = formattedTime
				}
			}
			if !prefixExists {
				HashStruct.NTLMHashPrefix = append(HashStruct.NTLMHashPrefix, prefix)
			}
			//HashStruct.NTLMHashPrefix = append(HashStruct.NTLMHashPrefix, prefix)
		}
	}

	if foundHash && foundHashOffline {
		return true, nil, true
	} else if foundHash && !foundHashOffline {
		return true, nil, false
	}

	return false, nil, false
}

func LoadOfflineFiles(c Configuration, hashType string) {
	var hashFiles []string
	var hashDirectory string
	if hashType == "SHA1" {
		hashDirectory = c.OfflineFiles + "/sha1/"
	} else {
		hashDirectory = c.OfflineFiles + "/ntlm/"
	}
	err := filepath.Walk(hashDirectory, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("[E] error accessing path %q: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		hashFiles = append(hashFiles, path)

		return nil
	})
	// Handle any errors from filepath.Walk
	if err != nil {
		log.Printf("[W] Warning walking the directory: %v", err)
	}

	for _, f := range hashFiles {
		var prefix PrefixStruct
		offlineFile, err := os.Open(f)
		if err != nil {
			log.Printf("[W] Warning reading an offlinefile: %v", err)
		}
		defer offlineFile.Close()
		decoder := json.NewDecoder(offlineFile)
		if err := decoder.Decode(&prefix); err != nil {
			log.Printf("[W] Unable to decode the JSON from offline file: %v", err)
		}
		if hashType == "SHA1" {
			HashStruct.SHA1HashPrefix = append(HashStruct.SHA1HashPrefix, prefix)
		} else {
			HashStruct.NTLMHashPrefix = append(HashStruct.NTLMHashPrefix, prefix)
		}
	}
}

// create the struct where the hashes are stored to be global
var HashStruct HashOfflineLookupStruct

func main() {
	var config Configuration
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load")
	InputPtr := flag.Bool("i", false, "Use Interactive Mode")
	ReadFilePtr := flag.String("f", "", "File to load and read line-by-line that contains SHA1 or NTLM hashes")
	SHA1FilePtr := flag.String("sha1", "", "File to load and read plain-text passwords and convert into SHA1 hashes")
	NTLMFilePtr := flag.String("ntlm", "", "File to load and read plain-text passwords and convert into NTLM hashes")
	flag.Parse()

	log.Println("Loading the following config file: " + *ConfigPtr + "\n")
	if err := config.LoadConfig(*ConfigPtr); err != nil {
		//fmt.Println("Could not load the configuration file, creating a new default config.json")
		config.CreateConfig()
		log.Fatalf("Modify the config.json file to customize how the tool functions: %v\n", err)
	}

	// Create the offlineFiles location specified in the config.json if they do not exist
	cf.CreateDirectory("/" + config.OfflineFiles)
	cf.CreateDirectory("/" + config.OfflineFiles + "/sha1")
	cf.CreateDirectory("/" + config.OfflineFiles + "/ntlm")

	// Load the offline files into the struct
	if !config.SkipLoadOfflineFiles {
		LoadOfflineFiles(config, "SHA1")
		LoadOfflineFiles(config, "NTLM")
	}

	//fmt.Println(HashStruct)

	var inputHashes []string

	if *InputPtr {
		inputHashes = append(inputHashes, InputFromStdin())
	} else if len(*ReadFilePtr) > 0 {
		inputHashes = InputFromFile(*ReadFilePtr, "hashes")
	} else if len(*SHA1FilePtr) > 0 {
		inputHashes = InputFromFile(*SHA1FilePtr, "sha1passwords")
	} else if len(*NTLMFilePtr) > 0 {
		inputHashes = InputFromFile(*NTLMFilePtr, "ntlmpasswords")
	}

	// Load offline hash files

	for _, hashInput := range inputHashes {
		// If the hash is not upper-case the offline storage stores the prefix in lower-case...
		hashInput = strings.ToUpper(hashInput)
		//log.Println(hashInput)
		pwned, err, foundHashOffline := CheckHash(hashInput, config)
		if err != nil {
			fmt.Printf("Error checking password: %v\n", err)
			return
		}

		if pwned && foundHashOffline {
			fmt.Printf("[+] Password Hash Exists in Offline Files: %s\n", hashInput)
		} else if pwned && !foundHashOffline {
			fmt.Printf("[+] Password Hash Exists in HIBP API: %s\n", hashInput)
		} else {
			fmt.Printf("[-] Password Hash Not Available: %s\n", hashInput)
		}
	}

	if len(inputHashes) > 0 {
		fmt.Println("\nCompleted the analysis...")
		// Save new offline database struct
		if !config.SkipSaveOfflineFiles {
			HashStruct.CreateOfflineFiles(config)
		}
	} else {
		fmt.Println("\nAnalysis is complete, however no hashes analyzed!")
		flag.Usage()
	}
}
