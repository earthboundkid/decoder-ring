package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

type modeFunc = func([]byte) ([]byte, error)

func main() {
	encode := flag.Bool("encode", false, "encode rather than decode")
	flag.BoolVar(encode, "e", false, "shortcut for -encode")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of decoder-ring:

    decoder-ring [-encode] <MODE>

MODE choices are %s.

`, getModes())
		flag.PrintDefaults()
	}
	flag.Parse()

	mode := modes[flag.Arg(0)].decoder
	if *encode {
		mode = modes[flag.Arg(0)].encoder
	}

	if flag.NArg() != 1 || mode == nil {
		flag.Usage()
		os.Exit(2)
	}

	if err := exec(mode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getModes() string {
	modesStr := make([]string, 0, len(modes))
	for mode := range modes {
		modesStr = append(modesStr, mode)
	}
	sort.Strings(modesStr)
	return strings.Join(modesStr, ", ")
}

func exec(f modeFunc) error {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	// Strip trailing newlines
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	b, err = f(b)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, io.MultiReader(
		bytes.NewReader(b),
		strings.NewReader("\n"),
	))
	return err
}

var modes = map[string]struct{ decoder, encoder modeFunc }{
	"hex":        {hexDec, hexEnc},
	"base64":     {base64Dec, base64Enc},
	"base64-url": {base64URLDec, base64URLEnc},
}

func hexEnc(src []byte) (dst []byte, err error) {
	dst = make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return
}

func hexDec(src []byte) ([]byte, error) {
	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	return dst[:n], err
}

func base64Enc(src []byte) (dst []byte, err error) {
	dst = make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return
}

func base64Dec(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, src)
	return dst[:n], err
}

func base64URLEnc(src []byte) (dst []byte, err error) {
	dst = make([]byte, base64.URLEncoding.EncodedLen(len(src)))
	base64.URLEncoding.Encode(dst, src)
	return
}

func base64URLDec(src []byte) ([]byte, error) {
	dst := make([]byte, base64.URLEncoding.DecodedLen(len(src)))
	n, err := base64.URLEncoding.Decode(dst, src)
	return dst[:n], err
}
