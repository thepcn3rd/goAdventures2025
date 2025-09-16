# Docker Adventures

I am using docker in a few locations on my network.  I am currently running, guacamole, revshells, CyberChef, and Firefox in a docker.  Each of these docker containers is fronted by a reverse TLS Proxy with authentication.  The project [here](/sslReverseProxy/README.md) is what I am using for the reverse proxy.

Notes about the configurations of these dockers:

The guacamole I modified to remove the port forwarding and the proxy routes to the docker IP Address of the guacd server.

The revshells and CyberChef I use the docker files that they provide again removing the port forwarding to then be done by the proxy.

## Firefox in a Docker
The purpose of this docker is to visit websites for the purposes of testing malware and all other things cyber security.  The docker image is reloaded by crontab every 24 hours.  I also modified the docker command to not port forward, use a volume and the proxy.  The volume is for placing files or downloading files and then pulling them into the environment.  This docker runs on an isolated host that is updated and rebooted daily.

![crontab image](/picts/crontabDockerAdv.png)
