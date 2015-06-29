# logspout-loggly
Logspout container for Docker and Loggly

[This is the suggested pattern from logspout](https://github.com/gliderlabs/logspout/tree/master/custom)

## How to use
docker -run -e LOGGLY_TOKEN=<token> -v /var/run/docker.sock:/tmp/docker.sock
