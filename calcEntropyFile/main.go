package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"flag"
	"fmt"
	"hash"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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


// Modifications:
// Change the output to work on windows, create an output file...
// Output the file metadata, signer, version, etc. Also...
// Create summary at the bottom of the output
// Have an option to disable the output of the chunks with high entropy...

*/

// Function to calculate the entropy of a byte slice
func calculateEntropy(data []byte) float64 {
	// Initialize a map to store the frequency of each byte
	frequency := make(map[byte]int)
	for _, b := range data {
		frequency[b]++
	}

	// Calculate the entropy
	var entropy float64
	dataLen := float64(len(data))
	for _, count := range frequency {
		probability := float64(count) / dataLen
		entropy -= probability * math.Log2(probability)
	}

	return entropy
}

func identifyMIMEType(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	// Identify the type of file by MIME Type
	// Only the first 512 bytes are used to sniff the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		panic(err)
	}

	contentType := http.DetectContentType(buffer)
	return contentType
}

func HashFile(filePath string, algo string) string {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return ""
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
		return ""
	}

	// Calculate the hash sum
	hashSum := hash.Sum(nil)

	// Convert the hash to a hex string
	hashString := fmt.Sprintf("%x", hashSum)

	return hashString
}

func addFileInformation(f string) {
	// Basic file info available on all platforms
	fileInfo, err := os.Stat(f)
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		return
	}

	fmt.Printf("\n%sBasic File Information%s\n", colorBlue, colorReset)
	fmt.Println("----------------------")
	fmt.Printf("%sName:%s %s\n", colorGreen, colorReset, fileInfo.Name())
	fmt.Printf("%sSize:%s %d bytes\n", colorGreen, colorReset, fileInfo.Size())
	fmt.Printf("%sPermissions:%s %s\n", colorGreen, colorReset, fileInfo.Mode())
	fmt.Printf("%sLast Modified:%s %s\n", colorGreen, colorReset, fileInfo.ModTime())
	fmt.Printf("%sMD5:%s %s\n", colorGreen, colorReset, HashFile(f, "md5"))
	fmt.Printf("%sSHA1:%s %s\n", colorGreen, colorReset, HashFile(f, "sha1"))
	fmt.Printf("%sSHA256:%s %s\n", colorGreen, colorReset, HashFile(f, "sha256"))
	fmt.Printf("%sMIME Type:%s %s\n", colorGreen, colorReset, identifyMIMEType(f))

	// Output what the command line of running "file"
	if runtime.GOOS == "linux" {
		cmd := exec.Command("file", f)
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("%sLinux \"file\" command output:%s %s", colorGreen, colorReset, string(output))
	}

}

var colorReset = "\033[0m"

const (
	colorRed           = "\033[31m" // Red
	colorGreen         = "\033[32m" // Green
	colorBlue          = "\033[34m" // Blue
	colorYellow        = "\033[33m" // Yellow
	colorMagenta       = "\033[35m" // Magenta (Purple)
	colorCyan          = "\033[36m" // Cyan
	colorWhite         = "\033[37m" // White
	colorBrightRed     = "\033[91m" // Bright Red
	colorBrightGreen   = "\033[92m" // Bright Green
	colorBrightYellow  = "\033[93m" // Bright Yellow
	colorBrightBlue    = "\033[94m" // Bright Blue
	colorBrightMagenta = "\033[95m" // Bright Magenta
	colorBrightCyan    = "\033[96m" // Bright Cyan
	colorBrightWhite   = "\033[97m" // Bright White
)

func main() {
	var chunkASCIIOutput bool
	filePtr := flag.String("f", "", "File to read and calculate entropy")
	chunkSizePtr := flag.Int("s", 256, "Size of Chunk Evaluated")
	flag.BoolVar(&chunkASCIIOutput, "d", false, "Disable Output of Chunk Information")
	flag.Parse()

	addFileInformation(*filePtr)

	file, err := os.Open(*filePtr)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error converting the file into a byte array")
		os.Exit(1)
	}

	entropy := calculateEntropy(data)
	fmt.Printf("\n%sEntropy of File:%s %.4f bits/byte", colorGreen, colorReset, entropy)
	// Low Entropy <= 5.0
	if entropy <= 5.0 {
		fmt.Printf(" %s(Low Entropy)%s\n", colorGreen, colorReset)
	} else if entropy > 6.5 {
		fmt.Printf(" %s(High Entropy)%s\n", colorRed, colorReset)
	} else {
		fmt.Printf(" %s(Medium Entropy)%s\n", colorYellow, colorReset)
	}

	// Calculate the chunks of a file...............................................
	chunkSize := *chunkSizePtr
	fmt.Printf("%sChunk size is set to:%s %d\n", colorGreen, colorReset, chunkSize)

	// Create a buffer to read 64 bytes at a time

	totalChunks := 0
	lowEntropyChunks := 0
	mediumEntropyChunks := 0
	highEntropyChunks := 0
	for i := 0; i < len(data); i += chunkSize {
		end := min(i+chunkSize, len(data))
		chunk := data[i:end]

		entropyChunk := calculateEntropy(chunk)

		// Low Entropy <= 5.0
		if entropyChunk <= 5.0 {
			lowEntropyChunks++
		} else if entropyChunk > 6.5 {
			highEntropyChunks++
		} else {
			mediumEntropyChunks++
		}
		// Medium Entropy 5.0 < Entropy <= 6.5
		// High Entropy > 6.5

		// Display the chunk if the entropy is higher than the base entropy...
		if entropyChunk > entropy {

			//if entropyChunk != entropy {
			if !chunkASCIIOutput {
				fmt.Printf("\n%sEntropy of chunk %d:%s %f", colorGreen, totalChunks, colorReset, entropyChunk)
				// Low Entropy <= 5.0
				if entropyChunk <= 5.0 {
					fmt.Printf(" %s(Low Entropy)%s\n", colorGreen, colorReset)
				} else if entropyChunk > 6.5 {
					fmt.Printf(" %s(High Entropy)%s\n", colorRed, colorReset)
				} else {
					fmt.Printf(" %s(Medium Entropy)%s\n", colorYellow, colorReset)
				}

				// Output the hex of the chunk
				var hexStr string
				//var hexStrArray []string
				var asciiStr string
				//var asciiStrArray []string
				for index, b := range chunk {
					hexStr += fmt.Sprintf("%02x ", b)
					if b >= 32 && b <= 126 { // printable ASCII range
						asciiStr += fmt.Sprintf("%c ", b)
					} else {
						asciiStr += ". " // fmt.Printf(".") // non-printable characters are replaced with a dot
					}
					if index%32 == 31 {
						fmt.Printf("%03d - %s | %s\n", index, hexStr, asciiStr)
						hexStr = ""
						//fmt.Printf("ASCII: %s\n", asciiStr)
						asciiStr = ""
					}

				}
			}

		}
		totalChunks++
	}

	fmt.Printf("\n%sTotal Chunks:%s %d\n", colorGreen, colorReset, totalChunks)
	fmt.Printf("%sChunks with Low:%s %d - %sMed:%s %d - %sHigh:%s %d %sEntropy%s\n\n", colorGreen, colorReset, lowEntropyChunks, colorYellow, colorReset, mediumEntropyChunks, colorRed, colorReset, highEntropyChunks, colorGreen, colorReset)

}
