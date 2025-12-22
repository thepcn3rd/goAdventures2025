package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
)

func CreateDirectory(createDir string) {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to get the working directory", err)
	}
	newDir := currentDir + createDir
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(newDir, os.ModePerm)
		if err != nil {
			log.Printf("Unable to create directory "+createDir, err)
		}
	}
}

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		log.Printf("unable to save to txt file: %v", err)
		os.Exit(0)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func CalcSHA256Hash(message string) string {
	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:])
}
