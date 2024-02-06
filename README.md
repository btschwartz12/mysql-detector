# MySQL Version Detector

This tool is designed to initiate a MySQL handshake with a given host and port, attempting to extract the MySQL version running, if any. This hopefully can be built upon to include more advanced fingerprinting and detection techniques, but for now, it's relatively basic.

The MySQL Version Detector be run either as a web server or a command-line tool.

## Quickstart

Navigate to [btschwartz.com/scanner](https://btschwartz.com/scanner) and give it a try with a host and port of your choice, or use machines I set up for testing (see below). Please be respectful:

- `107.173.251.101:3306`
- `193.42.60.203:4100`

Alternatively, you can run the scanner locally using the instructions below.

## Usage

Make sure you have Go installed on your system. If not, you can download it from the [official website](https://golang.org/dl/).

```bash
$ go version
go version go1.21.6 linux/amd64
```

To run as a command-line tool:

```bash
$ go run . cli <host> <port> # the host and port to scan
```

To run as a web server:

```bash
$ go run . web <listen-port> [url-prefix]
```

## Testing

There are some tests included in `scanner_test.go`. These are not exhaustive, and manual testing should be performed to ensure the tool works as expected. To run the tests, use the following command:

```bash
$ go test -v
```

To manually test the tool, I have exposed a MySQL instance on two servers I manage. You are free to use these for testing purposes, but please be respectful:

- `107.173.251.101:3306`
- `193.42.60.203:4100`
- *Both of these instances have no data, but if you really want to try to get in, be my guest!*

Here is how you can use the MySQL Version Detector to scan these servers:

```bash
$ go run . cli 107.173.251.101 3306
MySQL server version: 8.0.33-0ubuntu0.20.04.2
```

```bash
$ go run . cli 193.42.60.203 4100                    
MySQL server version: 8.0.36-0ubuntu0.22.04.1
````