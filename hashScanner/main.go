package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

/**

References: https://github.com/psypanda/hashID/
References: https://github.com/SmeegeSec/HashTag/tree/master # For password hashes ... openssl passwd -6

I came across a scenario where I needed to scan a file for hashes and report on them.  hashID
works great to identify a hash but needed to change the logic to scan through a file where the
hash could appear in various locations on a line.

Attempted the following to do what I needed...
gitleaks worked well however has issues identifying hashes (Worked better than trufflehog)
trufflehog is awesome how it will validate what it finds

Enhancements
1. (Done) Create a template that will allow new protoypes to be added
2. (Done) Take the original prototypes from the hashID github project and modify for scanning within a line
3. (Done) Resave the original prototypes with the modifications that need to be introduced
3a. (Done) Ability to enable or disable the detection of a hash
4. (Done) Build an exclusion regex list for strings that are commonly matched and should be excluded
5. (Done) Add SSH public key regex
6. (Done) Add max file size on files that are searched
7. (Done) Determine if the file has non-printable characters in the first 512 bytes

The new hashes that I have added include (Mod) and then in the notes I placed a date of when I added or modified them...

**/

type ExclusionsStruct struct {
	MaxFileSize  int      `json:"maxFileSize"`
	BinaryCheck  int      `json:"binaryCheck"` // The number of bytes to check to see if the file has non-printable ASCII characters
	MatchStrings []string `json:"matchStrings,omitempty"`
	Regexs       []string `json:"regexs,omitempty"`
	Files        []string `json:"files,omitempty"`
}

func (e *ExclusionsStruct) LoadFile(ePtr string) error {
	exclusionsFile, err := os.Open(ePtr)
	if err != nil {
		return err
	}
	defer exclusionsFile.Close()
	decoder := json.NewDecoder(exclusionsFile)
	if err := decoder.Decode(&e); err != nil {
		return err
	}

	return nil
}

func (e *ExclusionsStruct) CreateFile(f string) error {
	e.MaxFileSize = 1048576
	e.BinaryCheck = 512
	e.MatchStrings = append(e.MatchStrings, "string1")
	e.MatchStrings = append(e.MatchStrings, "string2")
	e.Regexs = append(e.Regexs, "^123$")
	e.Regexs = append(e.Regexs, "^abc$")
	e.Files = append(e.Files, "file1")
	e.Files = append(e.Files, "file2")

	jsonData, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Modified the original prototypes.json file to have "{ "prototypes": in the beginning...
// Added a } at the end to close the json...
type PrototypesStruct struct {
	Prototypes []PrototypeStruct `json:"prototypes"`
}

type PrototypeStruct struct {
	OriginalRegex string       `json:"regex"`
	NewRegex      string       `json:"newRegex,omitempty"`
	Enabled       bool         `json:"enabled,omitempty"`
	Notes         string       `json:"notes,omitempty"`
	Modes         []ModeStruct `json:"modes"`
}

type ModeStruct struct {
	John     string `json:"john"`
	Hashcat  int    `json:"hashcat"`
	Extended bool   `json:"extended"`
	HashName string `json:"name"`
}

func (ps *PrototypesStruct) CreateTemplateFile(f string) error {
	var p PrototypeStruct
	ps.Prototypes = append(ps.Prototypes, p)
	ps.Prototypes[0].Enabled = true
	ps.Prototypes[0].Notes = "Custom Template"
	ps.Prototypes[0].OriginalRegex = "^[123]+$"
	ps.Prototypes[0].NewRegex = "[123]+"
	var m ModeStruct
	ps.Prototypes[0].Modes = append(ps.Prototypes[0].Modes, m)
	ps.Prototypes[0].Modes[0].Extended = false
	ps.Prototypes[0].Modes[0].HashName = "Template"
	ps.Prototypes[0].Modes[0].John = ""
	ps.Prototypes[0].Modes[0].Hashcat = 0

	jsonData, err := json.MarshalIndent(ps, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (p *PrototypesStruct) LoadFile(pPtr string) error {
	configFile, err := os.Open(pPtr)
	if err != nil {
		return err
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&p); err != nil {
		return err
	}

	return nil
}

// This function exists to take the prototypes.json file from the hashID github project and convert it to the format we need in this project
func (ps *PrototypesStruct) SaveNewFile(oPtr string) error {
	for i, p := range ps.Prototypes {
		ps.Prototypes[i].Enabled = true
		ps.Prototypes[i].NewRegex = p.OriginalRegex[1 : len(p.OriginalRegex)-1]
		ps.Prototypes[i].Notes = "_"
	}

	jsonData, err := json.MarshalIndent(ps, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(oPtr, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func EvaluateRegex(pattern string, line string, p PrototypeStruct, e ExclusionsStruct, f string) {
	colorReset := "\033[0m"
	colorGreen := "\033[32m"
	colorBlue := "\033[34m"
	colorMaroon := "\033[38;5;88m"

	// Compile the regular expression
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println(p.OriginalRegex)
		log.Fatalln("[E] Error compiling regex:", err)
	}

	// Find the matching string with the patternValidate
	match := re.FindString(line)
	//if match != "" && len(match) == len(line) {
	//fmt.Println(e)
	// Is the match in the exclusion struct
	// The slices module is not avaiable in golang 1.18
	//if slices.Contains(e.MatchStrings, match) {
	//	return
	//}
	for _, e := range e.MatchStrings {
		if match == e {
			return
		}
	}
	for _, exclusionRegex := range e.Regexs {
		reExclusion, errE := regexp.Compile(exclusionRegex)
		if errE != nil {
			fmt.Println(exclusionRegex)
			log.Fatalln("[E] Error compiling regex:", err)
		}
		matchExclusion := reExclusion.FindString(match)
		if matchExclusion != "" && len(matchExclusion) == len(match) {
			return
		}
	}

	if match != "" {

		fmt.Printf("\n[*] Processing File: %s\n", f)
		fmt.Printf("%s[$] Original string%s: %s\n", colorBlue, colorReset, line)
		fmt.Printf("%s[$] Regex%s: %s\n", colorBlue, colorReset, p.OriginalRegex)
		fmt.Printf("%s[$] Matched on this string%s: %s\n", colorBlue, colorReset, match)

		for _, m := range p.Modes {
			fmt.Printf("%s[+]%s %s", colorGreen, colorReset, m.HashName)
			if m.John != "" {
				fmt.Printf("\t%sJohn:%s %s", colorBlue, colorReset, m.John)
			}
			if m.Hashcat != 0 {
				fmt.Printf("\t%sHashcat:%s %d", colorMaroon, colorReset, m.Hashcat)
			}
			fmt.Printf("\n")
		}
	}

}

func EvaluateFuzzyRegex(pattern string, line string, p PrototypeStruct, e ExclusionsStruct, f string) {
	colorReset := "\033[0m"
	colorGreen := "\033[32m"
	colorBlue := "\033[34m"
	colorMaroon := "\033[38;5;88m"
	// Characters to add before and after the string that matches the patter
	prefixRegex := "[\\s\\=\\x22\\x27\\x28\\x3c]{1}"
	sufixRegex := "([\\s\\x22\\x27\\x29\\x3e]{1}|$)"
	//var patternValidate string

	//fmt.Printf("Adding the following prefix regex: %s\tSufix: %s\n", prefixRegex, sufixRegex)
	patternValidate := prefixRegex + pattern + sufixRegex

	// Compile the regular expression
	reValidate, err := regexp.Compile(patternValidate)
	if err != nil {
		fmt.Println(p.OriginalRegex)
		log.Fatalln("[E] Error compiling regex:", err)
	}

	var match string
	// Find the matching string with the patternValidate
	matchValidation := reValidate.FindString(line)

	if matchValidation != "" {
		// Compile the regular expression
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Println(p.OriginalRegex)
			log.Fatalln("[E] Error compiling regex:", err)
		}
		match = re.FindString(line)
	}
	//if match != "" && len(match) == len(line) {
	// Is the match in the exclusion struct
	//if slices.Contains(e.MatchStrings, match) {
	//	return
	//}
	for _, e := range e.MatchStrings {
		if match == e {
			return
		}
	}
	for _, exclusionRegex := range e.Regexs {
		reExclusion, errE := regexp.Compile(exclusionRegex)
		if errE != nil {
			fmt.Println(exclusionRegex)
			log.Fatalln("[E] Error compiling regex:", err)
		}
		matchExclusion := reExclusion.FindString(match)
		if matchExclusion != "" && len(matchExclusion) == len(match) {
			return
		}
	}

	if match != "" {
		fmt.Printf("\n[*] Processing File: %s\n", f)
		fmt.Printf("%s[$] Original string%s: %s\n", colorBlue, colorReset, line)
		fmt.Printf("%s[$] Regex%s: %s\n", colorBlue, colorReset, p.NewRegex)
		fmt.Printf("%s[$] Adding the following prefix regex%s: %s\tSufix: %s\n", colorBlue, colorReset, prefixRegex, sufixRegex)
		fmt.Printf("%s[$] Matched on this string%s: %s\n", colorBlue, colorReset, match)
		for _, m := range p.Modes {
			fmt.Printf("%s[+]%s %s", colorGreen, colorReset, m.HashName)
			if m.John != "" {
				fmt.Printf("\t%sJohn:%s %s", colorBlue, colorReset, m.John)
			}
			if m.Hashcat != 0 {
				fmt.Printf("\t%sHashcat:%s %d", colorMaroon, colorReset, m.Hashcat)
			}
			fmt.Printf("\n")
		}
	}

}

func InputFromStdin(ps PrototypesStruct, e ExclusionsStruct, f string) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter text to analyze below:")
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("[E] Error reading input:", err)
	}
	fmt.Println()
	line = strings.Replace(line, "\r", "", -1)
	line = strings.Replace(line, "\n", "", -1)

	// Evaluate the original regex
	for _, p := range ps.Prototypes {
		// Evaluate the Original Regex
		EvaluateRegex(p.OriginalRegex, line, p, e, f)

	}
	// Evaluate the New Regex - This searches for the hash being present in a string with an ending delimeter of \n
	// Evaluating for the hash in the middle of a line...  This leads to a lot of false positives...
	//fmt.Printf("\n\n%s\n", strings.Repeat("*", 100))
	fmt.Printf("\n[*] Evaluating the regexs to find the hash in the middle of a series of characters\n")
	//fmt.Printf("%s\n", strings.Repeat("*", 100))
	for _, p := range ps.Prototypes {
		EvaluateFuzzyRegex(p.NewRegex, line, p, e, f)
	}
}

func InputFromFile(ps PrototypesStruct, f string, e ExclusionsStruct) {

	fmt.Printf("\n[*] Processing File: %s\n", f)

	for _, excludedFile := range e.Files {
		if f == excludedFile {
			fmt.Printf("[W] File is in an exclusion list: %s\n", f)
			return
		}
	}
	file, err := os.Open(f)
	if err != nil {
		log.Printf("[E] Failed to open file: %s", err)
	}
	defer file.Close()

	// Skip files that are binary - This function checks the first 512 bytes of the file to see if they are UTF-8 readable
	isBinary, _ := IsBinaryFile(f, e)
	if isBinary {
		fmt.Printf("[W] Binary File Detected: %s\n", f)
		return
	}

	fileInfo, _ := os.Stat(f)
	fileSize := fileInfo.Size()
	if fileSize > int64(e.MaxFileSize) {
		fmt.Printf("[W] File larger than max size: %d  Filename: %s\n", fileSize, f)
		return
	}

	// Create a new Scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read each line
	for scanner.Scan() {
		line := scanner.Text() // Get the current line as a string

		for _, p := range ps.Prototypes {
			// Only evaluate the prototypes that are enabled - This allows customization
			if p.Enabled {
				// Evaluate the Original Regex
				EvaluateRegex(p.OriginalRegex, line, p, e, f)
			}
		}
		// Evaluate the New Regex - This searches for the hash being present in a string with an ending delimeter of \n

		for _, p := range ps.Prototypes {
			// Only evaluate the prototypes that are enabled - This allows customization
			if p.Enabled {
				EvaluateFuzzyRegex(p.NewRegex, line, p, e, f)
			}
		}
	}

	// Check for any errors that occurred during scanning
	if err := scanner.Err(); err != nil {
		log.Printf("[E] Error reading file: %s", err)
	}
}

func InputFromFileChannels(ps PrototypesStruct, f string, e ExclusionsStruct, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{}
	defer func() { <-sem }()

	for _, excludedFile := range e.Files {
		if f == excludedFile {
			fmt.Printf("\n[*] Processing File: %s\n", f)
			fmt.Printf("[W] File is in an exclusion list: %s\n", f)
			return
		}
	}
	file, err := os.Open(f)
	if err != nil {
		log.Printf("[E] Failed to open file: %s", err)
	}
	defer file.Close()

	// Skip files that are binary - This function checks the first 512 bytes of the file to see if they are UTF-8 readable
	isBinary, _ := IsBinaryFile(f, e)
	if isBinary {
		fmt.Printf("\n[*] Processing File: %s\n", f)
		fmt.Printf("[W] Binary File Detected: %s\n", f)
		return
	}

	fileInfo, _ := os.Stat(f)
	fileSize := fileInfo.Size()
	if fileSize > int64(e.MaxFileSize) {
		fmt.Printf("\n[*] Processing File: %s\n", f)
		fmt.Printf("[W] File larger than max size: %d  Filename: %s\n", fileSize, f)
		return
	}

	// Create a new Scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read each line
	for scanner.Scan() {
		line := scanner.Text() // Get the current line as a string

		for _, p := range ps.Prototypes {
			// Only evaluate the prototypes that are enabled - This allows customization
			if p.Enabled {
				// Evaluate the Original Regex
				EvaluateRegex(p.OriginalRegex, line, p, e, f)
			}
		}
		// Evaluate the New Regex - This searches for the hash being present in a string with an ending delimeter of \n

		for _, p := range ps.Prototypes {
			// Only evaluate the prototypes that are enabled - This allows customization
			if p.Enabled {
				EvaluateFuzzyRegex(p.NewRegex, line, p, e, f)
			}
		}
	}

	// Check for any errors that occurred during scanning
	if err := scanner.Err(); err != nil {
		log.Printf("[E] Error reading file: %s", err)
	}
}

// Verify that the first 512 bytes are ascii readable text prior to processing the file line by line
func IsBinaryFile(filePath string, e ExclusionsStruct) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, e.BinaryCheck) // Read the first x bytes read from the binaryCheck key in the exclusions.json file
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	// Check for null bytes or non-printable characters
	if bytes.Contains(buffer[:n], []byte{0}) {
		return true, nil
	}

	for _, b := range buffer[:n] {
		// New Line is DEC 10
		// Carriage Return is DEC 13
		// Tab is DEC 9
		if b > 32 || b < 126 || b == 9 || b == 10 || b == 13 { // Non-printable ASCII characters
			return false, nil
		} else {
			return true, nil
		}
	}

	return false, nil
}

func main() {
	var prototypes PrototypesStruct
	// Rewrite the original to the new structure
	//OriginalPrototypesPtr := flag.String("o", "default.json", "Load the original prototypes file from the hashID project.")
	//NewPrototypesPtr := flag.String("n", "newDefault.json", "Save new prototypes file.")
	PrototypesPtr := flag.String("p", "customPrototypes.json", "Load a prototypes file to use")
	ExcludePtr := flag.String("e", "exclusions.json", "Loads a file that contains exclusions, or matched strings to ignore")
	InputPtr := flag.Bool("s", false, "Read from input to search for a match")
	FilePtr := flag.String("f", "", "Search the specified file for hashes")
	DirPtr := flag.String("d", "", "Search the specified directories and the files within for hashes")
	TemplatesPtr := flag.Bool("t", false, "Create a template prototypes.json file so I can create my own")
	flag.Parse()

	//var creatingNew bool
	//creatingNew = false
	/**
	// A value for prototypes needs to be specified or it does nothing to read the original and convert it
	if *OriginalPrototypesPtr != "default.json" {
		creatingNew = true
		log.Println("Loading the following original prototypes file: " + *OriginalPrototypesPtr + "\n")
		if err := prototypes.LoadFile(*OriginalPrototypesPtr); err != nil {
			//fmt.Println("Could not load the prototypes.json files")
			log.Fatalf("Could not load the prototypes.json files: %v\n", err)
		}
	}

	// A value for a new prototype file needs to be specified to save the file
	if *NewPrototypesPtr != "newDefault.json" {
		creatingNew = true
		log.Println("Saving a new prototypes file: " + *NewPrototypesPtr + "\n")
		if err := prototypes.SaveNewFile(*NewPrototypesPtr); err != nil {
			fmt.Printf("Could not create the %s files", *NewPrototypesPtr)
			log.Fatalf("Unable to save the new file: %v\n", err)
		}
	}
	**/

	// Create a template file in the event it is used
	if *TemplatesPtr {
		if !cf.FileExists("/" + "templatePrototypes.json") {
			log.Println("[*] Created templatePrototypes.json file for customization, exiting!")
			prototypes.CreateTemplateFile("templatePrototypes.json")
		}
	}

	// Load the exclusions file
	var exclusions ExclusionsStruct
	if cf.FileExists("/" + *ExcludePtr) {
		log.Println("[*] Loaded the following exclusions file: " + *ExcludePtr + "\n")
		err := exclusions.LoadFile(*ExcludePtr)
		if err != nil {
			log.Println("Failed to load the exclusions file: " + *ExcludePtr + "\n")
			os.Exit(0)
		}
	} else {
		log.Println("[E] Could not find exclusions.json, created it!")
		exclusions.CreateFile(*ExcludePtr)
	}

	prototypesLoaded := false

	log.Println("[*] Loading the following prototypes file: " + *PrototypesPtr + "\n")
	if err := prototypes.LoadFile(*PrototypesPtr); err != nil {
		//fmt.Printf("Could not load the %s file", *PrototypesPtr)
		log.Fatalf("[E] Could not load the %s file: %v\n", *PrototypesPtr, err)
	} else {
		prototypesLoaded = true
	}

	if *InputPtr && prototypesLoaded {
		InputFromStdin(prototypes, exclusions, "STDIN")
	} else if prototypesLoaded && len(*FilePtr) > 0 {
		InputFromFile(prototypes, *FilePtr, exclusions)
	} else if prototypesLoaded && len(*DirPtr) > 0 {
		//fmt.Println("Loaded...")
		// Walk through the directory recursively
		// This goes rather slow trying to improve performance with go threading...
		var filePaths []string
		err := filepath.Walk(*DirPtr, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return fmt.Errorf("[E] error accessing path %q: %w", path, err)
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Gather the filepaths of the files to parse
			filePaths = append(filePaths, path)

			return nil
		})

		// Handle any errors from filepath.Walk
		if err != nil {
			log.Printf("[W] Warning walking the directory: %v", err)
		}

		//lengthFilePaths := len(filePaths)
		var wg sync.WaitGroup
		// Process 10 files at a time...
		semaphore := make(chan struct{}, 10)
		for _, path := range filePaths {
			wg.Add(1)
			InputFromFileChannels(prototypes, path, exclusions, &wg, semaphore)
		}

		wg.Wait()
	}

}
