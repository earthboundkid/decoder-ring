package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

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

	funcs := decoders
	if *encode {
		funcs = encoders
	}
	mode := funcs[flag.Arg(0)]
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
	modes := make([]string, 0, len(decoders))
	for mode := range decoders {
		modes = append(modes, mode)
	}
	sort.Strings(modes)
	return strings.Join(modes, ",")
}

func exec(f func([]byte) ([]byte, error)) error {
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

var (
	decoders = map[string]func([]byte) ([]byte, error){
		"hex": hexDec,
	}
	encoders = map[string]func([]byte) ([]byte, error){
		"hex": hexEnc,
	}
)

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
