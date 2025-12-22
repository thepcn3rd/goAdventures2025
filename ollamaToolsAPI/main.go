package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		fmt.Printf("unable to write to txt file: %v", err)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func pullVirusTotalInfo(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	userAgent := r.UserAgent()
	fmt.Printf("IP Address: %s - User Agent: %s - Endpoint: /f1\n", ip, userAgent)
	w.Header().Set("Content-Type", "applicaton/json")
	ipAddr := r.URL.Query().Get("ipaddress")
	url := fmt.Sprintf("https://www.virustotal.com/api/v3/ip_addresses/%s", ipAddr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("x-apikey", "<APIKEY>")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "Error fetching VirusTotal information: %s", resp.Status)
	}
	ipInformation, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}
	fmt.Fprintf(w, "%s", string(ipInformation))
	fmt.Printf("VirusTotal information for IP: %s\n%s\n\n", ipAddr, string(ipInformation))

	SaveOutputFile(string(ipInformation), "virustotal_output.json")
}

func reverseDNSLookup(w http.ResponseWriter, r *http.Request) {
	dnsServer := "9.9.9.9"
	ip := r.RemoteAddr
	userAgent := r.UserAgent()
	fmt.Printf("IP Address: %s - User Agent: %s - Endpoint: /f3\n", ip, userAgent)
	w.Header().Set("Content-Type", "text/plain")
	ipAddr := r.URL.Query().Get("ipaddress")
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(dnsServer, "53"))
		},
	}

	// Perform the reverse DNS lookup
	names, err := resolver.LookupAddr(context.Background(), ipAddr)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	fmt.Fprintln(w, "Reverse DNS lookup results:")
	if len(names) > 0 {
		for _, name := range names {
			fmt.Fprintln(w, name)
		}
	} else {
		fmt.Fprintln(w, "No hostnames found")
	}
}

func pullGeoIPInfo(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	userAgent := r.UserAgent()
	fmt.Printf("IP Address: %s - User Agent: %s - Endpoint: /f2\n", ip, userAgent)
	w.Header().Set("Content-Type", "text/plain")
	ipAddr := r.URL.Query().Get("ipaddress")
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching IP information:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error fetching IP information:", resp.Status)
	}
	// Read the response body as a string
	ipInformation, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	// Write the response back to the client
	fmt.Fprintln(w, string(ipInformation))

}

func main() {
	http.HandleFunc("/f1", pullVirusTotalInfo)
	http.HandleFunc("/f2", pullGeoIPInfo)
	http.HandleFunc("/f3", reverseDNSLookup)

	fmt.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
