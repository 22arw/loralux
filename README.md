# loralux daemon

This daemon handles scraping a LoRaWAN server for lumen sensor data points. This repository also contains a test server
that can be used to test the loralux daemon.

## Table of Contents

- [Running](#running)
    - [Dependencies](#dependencies)
    - [Configuration](#environment-variables)
        - [From Environment](#from-environment)
        - [From File](#from-environment)
    - [Make Rules](#make-rules)

## Running

### Dependencies

The only dependencies to run the services in this repository are:

- `docker`
- `docker-compose`

### Configuration

The loralux daemon can be configured either via environment (default), or by file. If you want to configure the daemon
via a file, you must pass the `-envFile=` flag when running the daemon, giving it the path to either a JSON or YAML file
containing configuration values.

#### From Environment

The program looks for the following environment variables:

- `LORALUX_LOG_LEVEL`: The log level for the loralux daemon to use when logging, effectively filtering some statements.
See the [zap documentation](https://godoc.org/go.uber.org/zap/zapcore#Level) here for allowed levels (Default: `0`
(`InfoLevel`)).
- `LORALUX_SERVER_ADDRESS`: The server address of the LoRaWAN server to scrape (Default: `http://localhost:8080`).
- `LORALUX_SCRAPE_ENDPOINT`: The endpoint that should be used with the server address to scrape (Default: `/scrape`)
- `LORALUX_SCRAPE_INTERVAL`: The interval that the server and endpoint tandems will be scraped at (Default: `5s`).
- `LORALUX_READ_TIMEOUT`: The time of the read timeout of any outgoing read requests made by the internal HTTP server
(Default: `10s`).

If the environment variable has a supplied default and none are set within the context of the host machine, then the
default will be used.

#### From File

If you opt to set the daemon's configuration via a file, the following structures will need to be used for JSON or YAML,
respectively (in the following examples all fields will be set to their defaults, for granular descriptions of these
fields, see their equivalents in [From Environment](#from-environment)):

JSON:
```json
{
    "logLevel": 0,
    "serverAddress": "http://localhost:8080",
    "scrapeEndpoint": "/scrape",
    "scrapeInterval": "5s",
    "readTimeout": "10s",
}
```

YAML:
```yaml
logLevel: 0
serverAddress: http://localhost:8080
scrapeEndpoint: /scrape
scrapeInterval: 5s
readTimeout: 10s
```

### Make Rules

To run the services simply execute the following command:

```shell
make run
```

This will stop any containers defined by the compose file if already running and then rebuild the containers using the
compose file.

To retrieve the logs of the services:

```shell
make logs
```

To stop the services:

```shell
make stop
```

To remove all networks and volumes that the services have created (effectively allowing you to start from a clean state
upon the next run):

```shell
make down
```
