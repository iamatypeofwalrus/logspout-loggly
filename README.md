# logspout-loggly
Logspout container for Docker and Loggly

[This is the suggested pattern from logspout](https://github.com/gliderlabs/logspout/tree/master/custom)

## How to use

```sh
docker --run 'logspout-loggly'\
  -d \
  -e 'LOGGLY_TOKEN=<token>' \
  iamatypeofwalrus/logspout-loggly \
  --volume /var/run/docker.sock:/tmp/docker.sock
```
