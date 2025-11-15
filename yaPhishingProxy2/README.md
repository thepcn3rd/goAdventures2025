# Yet Another Phishing Proxy

This proxy is setup to emulate a man-in-the-middle which changes the URLs in transit.  The configuration file will need to be modified to determine what to proxy.

#### Configuration File (config.json)
The port and the sourceURL are what is listening for the connections.  The destination URL is what is being proxied and modified in the HTML.  You can also specify the server certificate and the server key to be used.  Inside the source code is information how to generate the certs using SSL if you would like a self-signed one.

```text
{
	"listeningPort": "443",
	"listeningURL": "example.proxy.local",
	"proxiedURL": "www.original.domain"
	"serverCert": "keys/server.crt"
	"serverKey": "keys/server.key"
}
```

#### Command Line Configuration
You can modify from the command line a different config.json file to utilize.
```text
Usage
  -config string
        Configuration file to load for the proxy (default "config.json")
```


![reflection.png](/yaPhishingProxy2/reflection.png)
