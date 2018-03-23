# decoder-ring [![GoDoc](https://godoc.org/github.com/carlmjohnson/decoder-ring?status.svg)](https://godoc.org/github.com/carlmjohnson/decoder-ring) [![Go Report Card](https://goreportcard.com/badge/github.com/carlmjohnson/decoder-ring)](https://goreportcard.com/report/github.com/carlmjohnson/decoder-ring)

Decoder-ring is a CLI tool for decoding/encoding from common formats.

## Installation

First install [Go](http://golang.org).

If you just want to install the binary to your current directory and don't care about the source code, run

```bash
GOBIN="$(pwd)" GOPATH="$(mktemp -d)" go get github.com/carlmjohnson/decoder-ring
```


## Screenshots
```bash
$ decoder-ring -h
Usage of decoder-ring:

    decoder-ring [-encode] <MODE>

MODE choices are base32, base32-crockford, base32-hex, base64, base64-url, go, hex, html, json, rot13, url-path, url-query, or an IANA encoding name.

  -e    shortcut for -encode
  -emit
        emit trailing newline (UTF-8) (default true)
  -encode
        encode rather than decode
  -s    shortcut for -strip (default true)
  -strip
        strip trailing newlines from input (default true)
  -t    shortcut for -emit (default true)


$ echo 'Hello, World!' | decoder-ring -e base64
SGVsbG8sIFdvcmxkIQ==
$ echo SGVsbG8sIFdvcmxkIQ== | decoder-ring base64
Hello, World!
$ echo 'Hello, World!' | decoder-ring rot13
Uryyb, Jbeyq!
$ echo 'Hello, World!' | decoder-ring ebcdic-cp-us | decoder-ring -e hex
c3a7c38125253fc28cc280c3af3fc38a25c380c281
```

## Endorsements

> Useful

— [barryzxb](https://www.reddit.com/r/golang/comments/86ewvx/decoderring_a_cli_tool_for_decodingencoding_from/dw4vmdy/)
