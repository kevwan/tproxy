# tproxy

English | [简体中文](readme-cn.md)

[![Go](https://github.com/kevwan/tproxy/workflows/Go/badge.svg?branch=main)](https://github.com/kevwan/tproxy/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kevwan/tproxy)](https://goreportcard.com/report/github.com/kevwan/tproxy)
[![Release](https://img.shields.io/github/v/release/kevwan/tproxy.svg?style=flat-square)](https://github.com/kevwan/tproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Why I wrote this tool

When I develop backend services and write [go-zero](https://github.com/zeromicro/go-zero), I often need to monitor the network traffic. For example:
1. monitoring gRPC connections, when to connect and when to reconnect
2. monitoring MySQL connection pools, how many connections and figure out the lifetime policy
3. monitoring any TCP connections on the fly

## Installation

```shell
$ go install github.com/kevwan/tproxy@latest
```

Or use docker images:

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1 ./tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

For arm64:

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1-arm64 ./tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

## Usages

```shell
$ tproxy --help
Usage of tproxy:
  -d duration
    	the delay to relay packets
  -l string
    	Local address to listen on (default "localhost")
  -p int
    	Local port to listen on
  -q	Quiet mode, only prints connection open/close and stats, default false
  -r string
    	Remote address (host:port) to connect
  -t string
    	The type of protocol, currently support grpc
```

## Examples

### Monitor gRPC connections

```shell
$ tproxy -p 8088 -r localhost:8081 -t grpc -d 100ms
```

- listen on localhost and port 8088
- redirect the traffic to `localhost:8081`
- protocol type to be gRPC
- delay 100ms for each packets

<img width="600" alt="image" src="https://user-images.githubusercontent.com/1918356/173970015-404dfab1-3de5-4937-89a5-3e4ddc53e5d6.png">

### Monitor MySQL connections

```shell
$ tproxy -p 3307 -r localhost:3306
```

<img width="600" alt="image" src="https://user-images.githubusercontent.com/1918356/173970130-944e4265-8ba6-4d2e-b091-1f6a5de81070.png">

## Give a Star! ⭐

If you like or are using this project, please give it a **star**. Thanks!
