# tproxy

[English](readme.md) | 简体中文

[![Go](https://github.com/kevwan/tproxy/workflows/Go/badge.svg?branch=main)](https://github.com/kevwan/tproxy/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kevwan/tproxy)](https://goreportcard.com/report/github.com/kevwan/tproxy)
[![Release](https://img.shields.io/github/v/release/kevwan/tproxy.svg?style=flat-square)](https://github.com/kevwan/tproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 为啥写这个工具

当我在做后端开发或者写 [go-zero](https://github.com/zeromicro/go-zero) 的时候，经常会需要监控网络连接，分析请求内容。比如：
1. 分析 gRPC 连接何时连接、何时重连
2. 分析 MySQL 连接池，当前多少连接，连接的生命周期是什么策略
3. 也可以用来观察和分析任何 TCP 连接

## 安装

```shell
$ GOPROXY=https://goproxy.cn/,direct go install github.com/kevwan/tproxy@latest
```

或者使用 docker 镜像：

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1 tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

arm64 系统:

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1-arm64 tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

## 用法

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

## 示例

### 分析 gRPC 连接

```shell
$ tproxy -p 8088 -r localhost:8081 -t grpc
```

- 侦听在 localhost 和 8088 端口
- 重定向请求到 `localhost:8081`
- 识别数据包格式为 gRPC
- 数据包延迟100毫秒

<img width="561" alt="image" src="https://user-images.githubusercontent.com/1918356/174335713-259de707-c9f5-40f1-a96f-780232d23f43.png">

### 分析 MySQL 连接

```shell
$ tproxy -p 3307 -r localhost:3306
```

<img width="600" alt="image" src="https://user-images.githubusercontent.com/1918356/173970130-944e4265-8ba6-4d2e-b091-1f6a5de81070.png">

## 欢迎 star！⭐

如果你正在使用或者觉得这个项目对你有帮助，请 **star** 支持，感谢！
