# twitch-bot-config
This is a micro-program to manage the bot configuration in Redis

# Overview

This is a simple application that will read data from a YAML file and load them as key-value pairs in Redis.

It will also monitor changes made to the file, and keep Redis up-to-date dynamically.

# Running as a stand-alone application

It is possible to run this application in stand-alone mode. The easiest way is to run the docker-compose pre-configured (see command below). This will start the application, a Redis instance, a Redis UI and also Jaeger (because we have some very small tracing).

# Running as part of a bigger solution

You can plug it into a bigger solution (which is how I use for my personal bot), just adding the necessary entries in your mais docker-compose.

```bash
docker-compose up -d
```

Use the docker-compose file from this repository and make sure that all the references (ports, hostnames, paths etc.) are correct for the new environment.

# Web UI addresses

- **Jaeger UI**: http://localhost:16686
- **Redis Commander**: http://localhost:6379
- **Kibana**: http://localhost:5601

# Ports used (exposed by the containers)

## Elasticsearch
- 9200
- 9300

## Fluent bit
- 24224/tcp
- 24224/udp
- 2020

## Kibana
- 5601

## Jaeger
- 5775/udp
- 6831/udp
- 6832/udp
- 5778
- 16686
- 14268
- 14250
- 9411

## Redis

## Redis Commander
- 8081
