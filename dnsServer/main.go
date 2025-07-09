package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	cf "github.com/thepcn3rd/goAdvsCommonFunctions"
)

var config Configuration

type Configuration struct {
	UpstreamDNS     string             `json:"upstreamDNS"`
	ServerBanner    string             `json:"serverBanner"`
	SyslogOptions   SyslogConfig       `json:"syslogOptions"`
	SaveFileOptions SaveFileConfig     `json:"saveFileOptions"`
	ARecords        []ARecordsStruct   `json:"aRecords"`
	TXTRecords      []TXTRecordsStruct `json:"txtRecords"`
}

type SyslogConfig struct {
	SyslogEnabled    string `json:"syslogEnabled"`
	SyslogServer     string `json:"syslogServer"`
	SyslogOriginName string `json:"syslogOriginName"`
}

type SaveFileConfig struct {
	SaveFileEnabled   string `json:"saveFileEnabled"`
	SaveFileBaseName  string `json:"saveFileBaseName"`
	SaveFileExtension string `json:"saveFileExtension"`
}

type ARecordsStruct struct {
	AName string `json:"aName"`
	IP    string `json:"ip"`
}

type TXTRecordsStruct struct {
	TXTName    string `json:"txtName"`
	TXTMessage string `json:"txtMessage"`
}

type CacheEntry struct {
	Response   *dns.Msg
	Expiration time.Time
}

type DNSServer struct {
	localRecords   map[string]string // Map of domain names to IP addresses
	reverseRecords map[string]string // Map of IP addresses to domain names
	txtRecords     map[string]string // Map of domain names to TXT records
	upstreamDNS    string
	cache          map[string]CacheEntry
	mutex          sync.RWMutex
}

func NewDNSServer(upstreamDNSString string) *DNSServer {
	aRecordsMap := make(map[string]string)
	for _, item := range config.ARecords {
		aRecordsMap[item.AName] = item.IP
		aNameItems := strings.Split(item.AName, ".") // Configures the A Record resolution to allow the lookup of www with the aName of www.site.name
		if len(aNameItems) > 1 {
			aRecordsMap[aNameItems[0]+"."] = item.IP
		}
	}

	ptrRecordsMap := make(map[string]string)
	for _, item := range config.ARecords {
		octets := strings.Split(item.IP, ".")
		reverseIPAddressString := fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa.", octets[3], octets[2], octets[1], octets[0])
		ptrRecordsMap[reverseIPAddressString] = item.AName
	}

	txtRecordsMap := make(map[string]string)
	for _, item := range config.TXTRecords {
		txtRecordsMap[item.TXTName] = item.TXTMessage
	}

	return &DNSServer{
		localRecords:   aRecordsMap,
		reverseRecords: ptrRecordsMap,
		txtRecords:     txtRecordsMap,
		upstreamDNS:    upstreamDNSString,
		cache:          make(map[string]CacheEntry),
	}
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		log.Println("Invalid query: no questions")
		return
	}

	question := r.Question[0]
	queryName := question.Name
	logMessage(fmt.Sprintf("Received query: %s", queryName), nil)

	s.mutex.RLock()
	cacheEntry, found := s.cache[queryName]
	s.mutex.RUnlock()

	if found && time.Now().Before(cacheEntry.Expiration) {
		logMessage(fmt.Sprintf("Cache hit for %s", queryName), nil)
		w.WriteMsg(cacheEntry.Response)
		return
	}

	if question.Qtype == dns.TypeA {
		s.mutex.RLock()
		localIP, found := s.localRecords[queryName]
		s.mutex.RUnlock()

		if found {
			logMessage(fmt.Sprintf("Serving local domain: %s -> %s", queryName, localIP), nil)
			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.Authoritative = true

			aRecord := &dns.A{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP(localIP),
			}
			msg.Answer = append(msg.Answer, aRecord)

			s.mutex.Lock()
			s.cache[queryName] = CacheEntry{
				Response:   msg,
				Expiration: time.Now().Add(60 * time.Second),
			}
			s.mutex.Unlock()

			w.WriteMsg(msg)
			return
		}
	} else if question.Qtype == dns.TypePTR {
		s.mutex.RLock()
		domainName, found := s.reverseRecords[queryName]
		s.mutex.RUnlock()

		if found {
			logMessage(fmt.Sprintf("Serving reverse lookup: %s -> %s", queryName, domainName), nil)
			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.Authoritative = true

			ptrRecord := &dns.PTR{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypePTR,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Ptr: domainName,
			}
			msg.Answer = append(msg.Answer, ptrRecord)

			s.mutex.Lock()
			s.cache[queryName] = CacheEntry{
				Response:   msg,
				Expiration: time.Now().Add(60 * time.Second),
			}
			s.mutex.Unlock()

			w.WriteMsg(msg)
			return
		}
	} else if question.Qtype == dns.TypeTXT {
		s.mutex.RLock()
		txtRecord, found := s.txtRecords[queryName]
		s.mutex.RUnlock()

		if found {
			logMessage(fmt.Sprintf("Serving TXT record: %s -> %s", queryName, txtRecord), nil)
			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.Authoritative = true

			txt := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{txtRecord},
			}
			msg.Answer = append(msg.Answer, txt)

			s.mutex.Lock()
			s.cache[queryName] = CacheEntry{
				Response:   msg,
				Expiration: time.Now().Add(60 * time.Second),
			}
			s.mutex.Unlock()

			w.WriteMsg(msg)
			return
		}
	}

	logMessage(fmt.Sprintf("Forwarding query for %s to upstream DNS %s", queryName, config.UpstreamDNS), nil)
	c := new(dns.Client)
	resp, _, err := c.Exchange(r, s.upstreamDNS)
	if err != nil {
		logMessage(fmt.Sprintln("Failed to forward query"), err)
		return
	}

	s.mutex.Lock()
	s.cache[queryName] = CacheEntry{
		Response:   resp,
		Expiration: time.Now().Add(60 * time.Second),
	}
	s.mutex.Unlock()

	w.WriteMsg(resp)
}

func logToFile(message string) {
	currentDate := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("logs/%s-%s%s", config.SaveFileOptions.SaveFileBaseName, currentDate, config.SaveFileOptions.SaveFileExtension)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(message)
}

func logToSyslog(message string) {
	logger, err := syslog.Dial("udp", config.SyslogOptions.SyslogServer, syslog.LOG_INFO|syslog.LOG_DAEMON, config.SyslogOptions.SyslogOriginName)
	if err != nil {
		log.Printf("Failed to connect to remote syslog server: %v", err)
		return
	}
	defer logger.Close()

	logger.Info(message)
}

func logMessage(message string, err error) {
	dnsServerBanner := config.ServerBanner
	if config.SyslogOptions.SyslogEnabled == "True" && err != nil {
		go logToSyslog(fmt.Sprintf("%s - %s: %v", dnsServerBanner, message, err))
	} else {
		go logToSyslog(fmt.Sprintf("%s - %s", dnsServerBanner, message))
	}

	if config.SaveFileOptions.SaveFileEnabled == "True" && err != nil {
		go logToFile(fmt.Sprintf("%s - %s: %v", dnsServerBanner, message, err))
	} else {
		go logToFile(fmt.Sprintf("%s - %s", dnsServerBanner, message))
	}

}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Load config.json file
	log.Println("Loading the following config file: " + *ConfigPtr + "\n")
	//go logToSyslog(fmt.Sprintf("Loading the following config file: %s\n", *ConfigPtr))
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)

	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	upstreamDNS := config.UpstreamDNS

	dnsServer := NewDNSServer(upstreamDNS)
	dns.HandleFunc(".", dnsServer.ServeDNS)

	server := &dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	cf.CreateDirectory("/logs")

	m := fmt.Sprintln("Starting DNS server on UDP 53")
	logMessage(m, nil)
	m = fmt.Sprintf("Upstream DNS Server: %s\n", config.UpstreamDNS)
	logMessage(m, nil)
	if err := server.ListenAndServe(); err != nil {
		//log.Fatalf("Failed to start DNS server: %v", err)
		logMessage("Failed to start DNS Server", err)
	}
}
