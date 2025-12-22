### Install Go

Setup Go on Linux (Manjaro) - To Be Created
```bash
sudo pacman -S go
```

Install the mingw tools for compiling in linux
```bash
pamac -S mingw-w64-gcc
```


### Setup the Environment 
Setup the environment for each program individually prior to compiling.  This verifies that the dependencies are located in the "src" folder and you are not pulling from outside of the projects folder.

```bash
go env -w GOROOT="/usr/lib/go"

go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter3/yaWebServer"

mkdir src  
```

NOTE: After I change the GOPATH in VSCode I will restart VSCode.  This reloads with the correct GOPATH for the project I am working on.  Then it will prompt to install a couple of packages.

Then I will copy the commonFunctions folder into the src directory.  This folder contains functions that are shared across projects.  Some of the functions in common.go do not allow compilation for windows projects, they can be removed.

After placing the files in the commonFunctions folder go into the directory and run the following command.  It creates a go.mod file
```bash
go mod init commonFunctions
```

I do not completely understand the following command, however it allows golang to read the src folder and fixes the complaint that commonFunctions is not in the std and shows the root path.
```bash
go env -w GO111MODULE='auto'
```

### Compile a Go Program in Linux and Windows

#### Linux with headers
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o progName.bin main.go
```

#### Linux without headers
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o progName.bin -ldflags "-w -s" main.go
```

NOTE: Added CGO_ENABLED=1 because if the binary passes to another OS with a different gcc version, the compiled binary will continue to work...  Manjaro was using glibc 2.32 2.34 when compiling, the parrot linux docker used 2.31 due to the difference CGO_ENABLED=0 will remove the dependency...

#### Windows without headers
```bash
GOOS=windows GOARCH=amd64 go build -o yaWebServer.exe -ldflags "-w -s" main.go
```

### Compile a Go Program to be a Windows DLL
Below is a go program that can be compiled to be a DLL.  Note 3 things that need to be changed:

1 - You need to add 'import "C"'' at the top near the imported libraries
2 - For the functions that will be exported for use place "//export Execfunc", Execfunc being the name of the function you will utilize
3 - The main function should be empty as it shows below, the Execfunc should contain what main used to conatin

```go
package main

import "C"
import os/exec

//export Execfunc
func Execfunc() {
	c := exec.Command("calc.exe")
	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {}
```

Execute the following command to create the DLL
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags="-w -s -H=windowsgui" -o newName.dll
```

To test the dll on Windows execute the following command with the name of the function you exported above.

```cmd
rundll32.exe newName.dll,Execfunc
```


### Compile a Go Program to become Shellcode for go-donut
Below is a go program that can be used to create shellcode.  

```go
package main

import os/exec

func main() {
	c := exec.Command("calc.exe")
	if err := c.Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}
```

Execute the following command to create the EXE
```bash
GOOS=windows GOARCH=amd64 go build -buildmode=pie -o newName.exe -ldflags "-w -s" main.go
```

Then create the shellcode using go-donut, go-donut allows you to compile and use it on linux

```cmd
./donut.bin -i newName.exe -o loader.bin
```

[Powershell Script to test the Injection of the Shellcode](https://github.com/thepcn3rd/goAdventures/blob/main/projects/fakeDataGenerator/memoryInject.ps1)




