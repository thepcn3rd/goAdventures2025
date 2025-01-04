# Adventures of Learning Go 2025

The following collection of golang programs are for educational use only.  In 2023 I started learning golang and have loved the journey.  Due to the size of the previous repository I have created a new one for this years projects.

# Operation 127.1 Lab
This is a collection of programs that I build in my home lab for a variety of purposes.  The purposes of these programs range across all spectrum's of cyber security.  Each program will have a brief explanation here and then they may possibly have more information on a dedicated page for the program. 

![Armored Scorpion](/picts/armoredScorpion.png)

[DNS Server](/dnsServer/README.md) - Simple DNS Server written in Golang to perform local lookups and then forward any additional requests to an upstream DNS server.  Designed it to be simple and to create a log or send the log by syslog to Elastic inside of a Security Onion Solution.

[SSL Reverse Proxy](/sslReverseProxy/README.md) - Setup a guacamole server by using docker to connect to multiple VM instances installed on a proxmox.  By default the guacamole server does not enable SSL, instead of modifying the docker image or setting up a nginx reverse proxy; I built this custom reverse proxy from a previous project.  This also allows logging to a local file or by syslog to Elastic.


# Additional Supporting Programs

[Compiling Go Notes](Compiling_Notes.md) - How to install, compile for Linux, Windows, a Windows DLL and a Windows Binary for Shellcode

[Setup git and obsidian](gitAndObsidian.md) Configure git and obsidian to work with these projects

[Configuration Notes for Kali 2024.2](Kali2024.md) Configuration notes for a fresh install of Kali Linux from ISO

## License

This project is licensed under the [MIT License](/LICENSE.md).


