package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/calcEntropy"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o calcEntropy.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o calcEntropy.exe -ldflags "-w -s" main.go

/*
References:

https://www.ibm.com/docs/en/qsip/7.5?topic=content-analyzing-files-embedded-malicious-activity
"...entropy is used as an indicator of the variability of bits per byte. Because each character in a data unit consists of 1 byte, the entropy value indicates the variation of the characters and the compressibility of the data unit. Variations in the entropy values in the file might indicate that suspect content is hidden in files. For example, the high entropy values might be an indication that the data is stored encrypted and compressed and the lower values might indicate that at runtime the payload is decrypted and stored in different sections. "


http://www.forensickb.com/2013/03/file-entropy-explained.html
"The equation used by Shannon has a resulting value of something between zero (0) and eight (8). The closer the number is to zero, the more orderly or non-random the data is. The closer the data is to the value of eight, the more random or non-uniform the data is."


Create a json or csv output from the information collected...

*/

type EntropyStructs struct {
	BaseDir      string
	ChunkSize    int
	MaxDepth     int
	Created      string
	FileList     []string
	EntropyFiles []EntropyFile
	Debug        bool
	MaxFileSize  int // Max file size to evaluate chunks
}

type EntropyFile struct {
	Name                string
	FilePath            string
	FileSize            int64
	Permissions         string
	LastModified        string
	MD5                 string
	SHA1                string
	SHA256              string
	MIMEType            string
	Entropy             float64
	EntropyRating       string
	ChunksEvaluated     bool
	ChunkSize           int
	TotalChunks         int
	LowEntropyChunks    int
	MediumEntropyChunks int
	HighEntropyChunks   int
}

func (e *EntropyFile) CalculateEntropy(d []byte, selection string) error {
	// Initialize a map to store the frequency of each byte
	frequency := make(map[byte]int)
	for _, b := range d {
		frequency[b]++
	}

	// Calculate the entropy
	var entropy float64
	dataLen := float64(len(d))
	for _, count := range frequency {
		probability := float64(count) / dataLen
		entropy -= probability * math.Log2(probability)
	}

	if selection == "file" {
		e.Entropy = entropy

		// Calculate the rating of the Entropy
		// Low Entropy <= 5.0
		if entropy <= 5.0 {
			e.EntropyRating = "Low"
		} else if entropy > 6.5 {
			e.EntropyRating = "High"
		} else {
			e.EntropyRating = "Medium"
		}
		return nil
	} else if selection == "chunk" {
		// Low Entropy <= 5.0
		if entropy <= 5.0 {
			e.LowEntropyChunks++
		} else if entropy > 6.5 {
			e.HighEntropyChunks++
		} else {
			e.MediumEntropyChunks++
		}
	}

	return nil
}

func (e *EntropyFile) AddFileInformation(f string) error {
	fileInfo, err := os.Stat(f)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return nil
	}

	e.Name = fileInfo.Name()
	e.FilePath = f
	e.FileSize = fileInfo.Size()
	e.Permissions = fileInfo.Mode().String()
	e.LastModified = fileInfo.ModTime().String()

	return nil
}

func (e *EntropyFile) CalculateHashes(f string) error {
	HashFile(f, "md5", &e.MD5)
	HashFile(f, "sha1", &e.SHA1)
	HashFile(f, "sha256", &e.SHA256)
	return nil
}

func (e *EntropyFile) IdentifyMIMEType(f string) error {
	file, err := os.Open(f)
	if err != nil {
		return fmt.Errorf("[W] Unable to open file to identify MIME type: %v", err)
	}
	defer file.Close()
	// Identify the type of file by MIME Type
	// Only the first 512 bytes are used to sniff the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("[W] Unable to read the buffer to identify MIME type: %v", err)
	}

	e.MIMEType = http.DetectContentType(buffer)
	return nil
}

func Debug(message string, debug bool) {
	// Could save to a file if configured...
	if debug {
		fmt.Print(message)
	}
}

func (e *EntropyStructs) CreateJSONFile(f string) error {
	jsonData, err := json.MarshalIndent(e, "", "   ")
	if err != nil {
		return err
	}
	filename := f + ".json"
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (e *EntropyStructs) CreatCSVFile(f string) error {
	// Create a new CSV file
	file, err := os.Create(f + ".csv")
	if err != nil {
		return fmt.Errorf("failed to create the csv file: %v", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Name",
		"FilePath",
		"FileSize",
		"Permissions",
		"LastModified",
		"MD5",
		"SHA1",
		"SHA256",
		"MIMEType",
		"Entropy",
		"EntropyRating",
		"ChunksEvaluated",
		"ChunkSize",
		"TotalChunks",
		"LowEntropyChunks",
		"MediumEntropyChunks",
		"HighEntropyChunks",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing the csv header failed: %v", err)
	}

	// Write data rows
	for _, eFile := range e.EntropyFiles {
		record := []string{
			eFile.Name,
			eFile.FilePath,
			strconv.FormatInt(eFile.FileSize, 10),
			eFile.Permissions,
			eFile.LastModified,
			eFile.MD5,
			eFile.SHA1,
			eFile.SHA256,
			RemoveBadChars(eFile.MIMEType),
			fmt.Sprintf("%.4f", eFile.Entropy),
			eFile.EntropyRating,
			strconv.FormatBool(eFile.ChunksEvaluated),
			strconv.Itoa(eFile.ChunkSize),
			strconv.Itoa(eFile.TotalChunks),
			strconv.Itoa(eFile.LowEntropyChunks),
			strconv.Itoa(eFile.MediumEntropyChunks),
			strconv.Itoa(eFile.HighEntropyChunks),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("writing the csv file failed: %v", err)
		}
	}
	return nil
}

func (e *EntropyStructs) GatherFileList() error {
	err := filepath.Walk(e.BaseDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("[E] Error accessing path %q: %w", path, err)
		}

		relPath, err := filepath.Rel(e.BaseDir, path)
		if err != nil {
			return fmt.Errorf("[E] Error identifying the relative path")
		}

		currentDepth := len(strings.Split(relPath, string(filepath.Separator)))

		// Skip if beyond max depth (but maxDepthPtr must be >= 0)
		if e.MaxDepth >= 0 && currentDepth > e.MaxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}
		e.FileList = append(e.FileList, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("[W] Warning walking the directory: %v", err)
	}

	return nil
}

func RemoveBadChars(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, ";", "", -1)
	s = strings.Replace(s, ",", "", -1)
	s = strings.Replace(s, " ", "_", -1)
	return s
}

func HashFile(filePath string, algo string, hashReturned *string) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Create a new SHA1 hash
	var hash hash.Hash
	switch algo {
	case "sha1":
		hash = sha1.New()
	case "md5":
		hash = md5.New()
	case "sha256":
		hash = sha256.New()
	default:
		hash = md5.New()
	}

	// Copy the file content to the hash calculator
	if _, err := io.Copy(hash, file); err != nil {
		return
	}

	// Calculate the hash sum
	hashSum := hash.Sum(nil)

	// Convert the hash to a hex string
	*hashReturned = fmt.Sprintf("%x", hashSum)

}

func calcTime() string {
	now := time.Now()
	return now.Format("15:04:05")
}

func main() {
	directoryPtr := flag.String("d", "", "Calculate Entropy of Files in Specified Directory")
	outputPtr := flag.String("o", "output", "Save output to this file, extension depends on the format selected")
	formatPtr := flag.String("format", "both", "Output in CSV and JSON Format or specify")
	chunkSizePtr := flag.Int("size", 256, "Size of Chunk Evaluated")
	maxDepthPtr := flag.Int("depth", -1, "Maximum recursion depth (0 for current directory only, -1 for unlimited)")
	maxSizePtr := flag.Int("maxsize", 10, "Maximum size of file in MB to evaluate chunks")
	debugPtr := flag.Bool("debug", false, "Enable Debug Information, creates a debug file")
	flag.Parse()

	var eStruct EntropyStructs
	//var eFile EntropyFile

	eStruct.BaseDir = *directoryPtr
	eStruct.ChunkSize = *chunkSizePtr
	eStruct.MaxDepth = *maxDepthPtr
	eStruct.Debug = *debugPtr
	eStruct.MaxFileSize = *maxSizePtr * 1024 * 1024

	// Gather File list
	baseDir := *directoryPtr
	if len(baseDir) > 0 {
		Debug(fmt.Sprintf("Calculating the Entropy of the Files in the Specified Directory: %s\n", baseDir), eStruct.Debug)
		err := eStruct.GatherFileList()
		if err != nil {
			log.Printf("Error in calculating entropy: %v", err)
		}
	} else {
		flag.Usage()
		log.Fatalf("[E] Specify the directory to get the entropy of the files\n")
	}

	Debug(fmt.Sprintf("Number of Files to be Evaluated: %d\n", len(eStruct.FileList)), eStruct.Debug)

	// Iterate over the files in the file list
	for _, f := range eStruct.FileList {
		Debug(fmt.Sprintf("\nEvaluating the file: %s - %s\n", f, calcTime()), eStruct.Debug)
		// Add information about a file to a struct
		var eFile EntropyFile
		eFile.AddFileInformation(f)

		// Calculate the Hashes
		Debug(fmt.Sprintf("- Calculating Hashes (MD5, SHA1, SHA256) - %s\n", calcTime()), eStruct.Debug)
		err := eFile.CalculateHashes(f)
		if err != nil {
			log.Printf("[W] Calculating Hashes: %v", err)
		}

		if eFile.SHA256 == "" {
			eFile.MIMEType = "Access is Denied"
			eFile.EntropyRating = "None"
		} else {

			// Identify MIME Type of File
			Debug(fmt.Sprintf("- Identifying the MIME Type - %s\n", calcTime()), eStruct.Debug)
			err = eFile.IdentifyMIMEType(f)
			if err != nil {
				log.Printf("[W] Identifying the MIME Type: %v", err)
			}

			// Calculate Entropy of the File
			file, err := os.Open(f)
			if err != nil {
				log.Println("Error opening file:", err)
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				log.Println("Error converting the file into a byte array")
			}

			Debug(fmt.Sprintf("- Calculating the Entropy of the File - %s\n", calcTime()), eStruct.Debug)
			eFile.CalculateEntropy(data, "file")

			// Calculate Entropy of Chunks
			if eFile.FileSize < int64(eStruct.MaxFileSize) {
				eFile.ChunksEvaluated = true
				Debug(fmt.Sprintf("- Calculating the Entropy of the Chunks in the File - %s\n", calcTime()), eStruct.Debug)
				eFile.LowEntropyChunks = 0
				eFile.MediumEntropyChunks = 0
				eFile.HighEntropyChunks = 0

				for i := 0; i < len(data); i += eStruct.ChunkSize {
					end := min(i+eStruct.ChunkSize, len(data))
					chunk := data[i:end]
					eFile.CalculateEntropy(chunk, "chunk")
				}
			} else {
				Debug(fmt.Sprintf("- Skipped calculating the Entropy of the Chunks due to size of file is larger than maxsize - %s\n", calcTime()), eStruct.Debug)
				eFile.ChunksEvaluated = false
				eFile.LowEntropyChunks = 0
				eFile.MediumEntropyChunks = 0
				eFile.HighEntropyChunks = 0
			}

		}

		// Append the information with the other structs
		eStruct.EntropyFiles = append(eStruct.EntropyFiles, eFile)
	}

	if *formatPtr == "json" {
		eStruct.CreateJSONFile(*outputPtr)
	} else if *formatPtr == "csv" {
		eStruct.CreatCSVFile(*outputPtr)
	} else {
		// Output in json the information collected
		eStruct.CreateJSONFile(*outputPtr)

		// Output in csv the information collected
		eStruct.CreatCSVFile(*outputPtr)
	}
}
