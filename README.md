Sketch Canvas
=============

[![Go Report Card](https://goreportcard.com/badge/github.com/hexbee-net/sketch-canvas)](https://goreportcard.com/report/github.com/hexbee-net/sketch-canvas)
[![Apache 2 License](https://img.shields.io/github/license/hexbee-net/sketch-canvas)](https://github.com/hexbee-net/sketch-canvas/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/go-v1.17+-green.svg?style=flat)](https://github.com/hexbee-net/sketch-canvas-dispatch)
[![test](https://github.com/hexbee-net/sketch-canvas/actions/workflows/test.yml/badge.svg)](https://github.com/hexbee-net/sketch-canvas/actions/workflows/test.yml)
[![golangci-lint](https://github.com/hexbee-net/sketch-canvas/actions/workflows/lint.yml/badge.svg)](https://github.com/hexbee-net/sketch-canvas/actions/workflows/lint.yml)

## Quick Start

You can use `docker-compose` to run the whole stack:

```bash
$ docker-compose up
```

The server will be available at http://localhost:8800. If the port is not available, you can customize the port by
editing the `canvas` command in `docker-compose.yml`.

You can check the [API documentation](index.html) for instruction on how to access the server.

### Postman

You can use [Postman](https://www.postman.com/) to test the API directly. Use the provided [postman collection](postman_collection.json) file to get
predefined queries to use with the server.

## Command line parameters

```
-s, --datastore string            the hostname of the redis datastore (default "localhost:6379")
    --datastore-db int            the database to be selected on the redis server
    --datastore-password string   the password of the redis server
    --debug                       debug mode
-w, --graceful-timeout duration   the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m (default 15s)
-p, --port int                    the port number of the canvas server (default 8800)
-v, --verbose                     verbose mode
```
