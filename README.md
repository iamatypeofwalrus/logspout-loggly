# logspout-loggly
Logspout container for Docker and Loggly.

[This repo follows this suggested pattern from logspout](https://github.com/gliderlabs/logspout/tree/master/custom)

## How to run

```sh
docker --run 'logspout-loggly'\
  -d \
  -e 'LOGGLY_TOKEN=<token>' \
  --volume /var/run/docker.sock:/tmp/docker.sock \
  iamatypeofwalrus/logspout-loggly
```

## What it does
Instead of linking containers together or having to bother with (syslog or remote syslog)[https://www.loggly.com/blog/centralize-logs-docker-containers] this container follows the (12 Factor app logging philosophy)[http://12factor.net/logs]. If your docker container logs to STDOUT this image will pick up that stream from the docker daemon send those logs to Loggly. This container will log the STDOUT from any container that the docker daemon is managing.
