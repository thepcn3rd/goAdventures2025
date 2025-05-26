package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	iofs "io/fs"
	"log"
	"net"
	"os"

	"github.com/hirochachacha/go-smb2"
	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

/**
Program to connect to a SMB share and download files from it to a specified directory.
The program will also delete files older than a specified number of days in the local directory.


**/

type Config struct {
	Server            string `json:"server"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"encrypted_password"`
	Domain            string `json:"domain"`
	ShareName         string `json:"share_name"`
	LocalFilePath     string `json:"local_file_path"`
	FileRetentionDays int    `json:"file_retention_days"`
}

func (c *Config) LoadFile(f string) error {
	configFile, err := os.Open(f)
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

func (c *Config) CreateFile(f string) error {
	c.Server = "10.10.10.10"
	c.Username = "bradly.mark.fred.smith.freeman.rice"
	c.Password = "Wonderful-Blocked-17_Yellow;14_Grass"
	c.EncryptedPassword = ""
	c.Domain = "local.local"
	c.ShareName = "files"
	c.LocalFilePath = "downloads"
	c.FileRetentionDays = 90

	jsonData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(f, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var fileList []string
	var config Config
	encryptionKey := "dbGtQcnSRomEc1cW4fXoCgxkjeyISPxG" // Can be seen in the binary file in plain text
	configPtr := flag.String("config", "config.json", "Path to the config file")
	flag.Parse()

	if err := config.LoadFile(*configPtr); err != nil {
		config.CreateFile("config.json")
		log.Fatalf("Error loading config file, created config.json\n%v\n", err)
	}

	// Create the local directory if it doesn't exist
	cf.CreateDirectory("/" + config.LocalFilePath)
	// Delete files older than 90 days
	cf.DeleteFilesOlderThan("./"+config.LocalFilePath+"/", config.FileRetentionDays) // Delete files older than 90 days

	salt, err := cf.GenerateSalt(".salt") // Reads the salt from the file or generates a new one
	if err != nil {
		log.Fatalf("Error Generating Information: %v\n", err)
	}

	// If the encrypted password is empty, encrypt the password
	if config.EncryptedPassword == "" {

		config.Password = config.Password + salt
		encryptedPassword, err := cf.EncryptStringGCM([]byte(encryptionKey), config.Password)
		if err != nil {
			log.Fatalf("Error encrypting password: %v\n", err)
		}
		fmt.Printf("\nReplace the encrypted password in the config file with the following and remove the password\n\n")
		fmt.Printf("Encrypted password: %s\n\n", encryptedPassword)
		os.Exit(0)
	}

	if config.Password != "" {
		fmt.Println("Remove the password that is in plaintext in the config file")
		os.Exit(0)
	}

	// Connect to the SMB server
	conn, err := net.Dial("tcp", config.Server+":445")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer conn.Close()

	decryptedPassword, err := cf.DecryptStringGCM([]byte(encryptionKey), config.EncryptedPassword)
	if err != nil {
		log.Fatalf("Error decrypting password: %v\n", err)
	}
	config.Password = string(decryptedPassword)
	config.Password = config.Password[:len(config.Password)-len(salt)]
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     config.Username,
			Password: config.Password,
			Domain:   config.Domain,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		fmt.Printf("Failed to dial SMB: %v\n", err)
		return
	}
	defer s.Logoff()

	// Mount the share
	fs, err := s.Mount(config.ShareName)
	if err != nil {
		fmt.Printf("Failed to mount share: %v\n", err)
		return
	}
	defer fs.Umount()
	// Walk through the files in the share
	err = iofs.WalkDir(fs.DirFS("."), ".", func(path string, d iofs.DirEntry, err error) error {
		//fmt.Println(path)
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
		}
		if d.IsDir() {
			//fmt.Printf("Directory: %s\n", path)
			return nil
		}

		fileList = append(fileList, path)

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
		return
	}

	// Open the remote files one at a time
	for _, file := range fileList {
		fmt.Println(file)
		remoteFile, err := fs.Open(file)
		if err != nil {
			fmt.Printf("Failed to open remote file: %v\n", err)
			return
		}
		defer remoteFile.Close()

		// Create the local file

		localFile, err := os.Create("./" + config.LocalFilePath + "/" + file)
		if err != nil {
			fmt.Printf("Failed to create local file: %v\n", err)
			return
		}
		defer localFile.Close()

		// Copy the file contents
		_, err = io.Copy(localFile, remoteFile)
		if err != nil {
			fmt.Printf("Failed to copy file: %v\n", err)
			return
		}

		fmt.Printf("File copied successfully from %s to %s\n", file, config.LocalFilePath)
	}
}
