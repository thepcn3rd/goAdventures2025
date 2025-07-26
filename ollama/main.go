package main

import (
	"fmt"
	"net/http"
)

// Function 1: Returns fake data if an IP Address is Malicious or Safe
func functionOneHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	userAgent := r.UserAgent()
	fmt.Printf("IP Address: %s - User Agent: %s\n", ip, userAgent)
	w.Header().Set("Content-Type", "text/plain")
	ipAddr := r.URL.Query().Get("ipaddress")
	if ipAddr == "5.5.5.5" {
		fmt.Fprintln(w, "Malicious")
	} else {
		fmt.Fprintln(w, "Safe")
	}
}

func main() {
	http.HandleFunc("/f1", functionOneHandler)

	fmt.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
