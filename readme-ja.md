# tproxy

[English](readme.md) | [简体中文](readme-cn.md) | 日本語

[![Go](https://github.com/kevwan/tproxy/workflows/Go/badge.svg?branch=main)](https://github.com/kevwan/tproxy/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kevwan/tproxy)](https://goreportcard.com/report/github.com/kevwan/tproxy)
[![Release](https://img.shields.io/github/v/release/kevwan/tproxy.svg?style=flat-square)](https://github.com/kevwan/tproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

<a href="https://www.buymeacoffee.com/kevwan" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>

## なぜこのツールを書いたのか

バックエンドサービスを開発し、[go-zero](https://github.com/zeromicro/go-zero)を書くとき、ネットワークトラフィックを監視する必要がよくあります。例えば：
1. gRPC接続の監視、接続のタイミングと再接続のタイミング
2. MySQL接続プールの監視、接続数とライフタイムポリシーの把握
3. 任意のTCP接続のリアルタイム監視

## インストール

```shell
$ go install github.com/kevwan/tproxy@latest
```

または、dockerイメージを使用します：

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1 tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

arm64の場合：

```shell
$ docker run --rm -it -p <listen-port>:<listen-port> -p <remote-port>:<remote-port> kevinwan/tproxy:v1-arm64 tproxy -l 0.0.0.0 -p <listen-port> -r host.docker.internal:<remote-port>
```

Windowsの場合、[scoop](https://scoop.sh/)を使用できます：

```shell
$ scoop install tproxy
```

## 使用方法

```shell
$ tproxy --help
Usage of tproxy:
  -d duration
    	パケットを中継する遅延時間
  -down int
    	下り速度制限（バイト/秒）
  -l string
    	リッスンするローカルアドレス（デフォルトは "localhost"）
  -p int
    	リッスンするローカルポート、デフォルトはランダムポート
  -q	静音モード、接続の開閉と統計のみを表示、デフォルトはfalse
  -r string
    	接続するリモートアドレス（ホスト：ポート）
  -s	統計を有効にする
  -t string
    	プロトコルの種類、現在サポートされているのはhttp2、grpc、redis、mongodb
  -up int
    	上り速度制限（バイト/秒）
```

## 例

### gRPC接続の監視

```shell
$ tproxy -p 8088 -r localhost:8081 -t grpc -d 100ms
```

- localhostとポート8088でリッスン
- トラフィックを`localhost:8081`にリダイレクト
- プロトコルタイプをgRPCに設定
- 各パケットの遅延時間を100msに設定

<img width="579" alt="image" src="https://user-images.githubusercontent.com/1918356/181794530-5b25f75f-0c1a-4477-8021-56946903830a.png">

### MySQL接続の監視

```shell
$ tproxy -p 3307 -r localhost:3306
```

<img width="600" alt="image" src="https://user-images.githubusercontent.com/1918356/173970130-944e4265-8ba6-4d2e-b091-1f6a5de81070.png">

### 接続の信頼性の確認（再送率とRTT）

```shell
$ tproxy -p 3307 -r remotehost:3306 -s -q
```

<img width="548" alt="image" src="https://user-images.githubusercontent.com/1918356/180252614-7cf4d1f9-9ba8-4aa4-a964-6f37cf991749.png">

### 接続プールの動作を学ぶ

```shell
$ tproxy -p 3307 -r localhost:3306 -q -s
```

<img width="404" alt="image" src="https://user-images.githubusercontent.com/1918356/236633144-9136e415-5763-4051-8c59-78ac363229ac.png">

## スターを付けてください！ ⭐

このプロジェクトが気に入ったり、使用している場合は、**スター**を付けてください。ありがとうございます！
