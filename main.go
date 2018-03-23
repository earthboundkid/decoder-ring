package main

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/ianaindex"
)

type modeFunc = func([]byte) ([]byte, error)

var modes = map[string]struct{ decoder, encoder modeFunc }{
	"hex":              {hexDec, hexEnc},
	"hex-extended":     {nil, hexExtEnc},
	"base32":           {base32Dec, base32Enc},
	"base32-crockford": {base32CrockfordDec, base32CrockfordEnc},
	"base32-hex":       {base32HexDec, base32HexEnc},
	"base64":           {base64Dec, base64Enc},
	"base64-url":       {base64URLDec, base64URLEnc},
	"go":               {goDec, goEnc},
	"html":             {htmlDec, htmlEnc},
	"json":             {jsonDec, jsonEnc},
	"rot13":            {rot13, rot13},
	"url-path":         {urlPathDec, urlPathEnc},
	"url-query":        {urlQueryDec, urlQueryEnc},
}

func main() {
	encode := flag.Bool("encode", false, "encode rather than decode")
	flag.BoolVar(encode, "e", false, "shortcut for -encode")
	strip := flag.Bool("strip", true, "strip trailing newlines from input")
	flag.BoolVar(strip, "s", true, "shortcut for -strip")
	emit := flag.Bool("emit", true, "emit trailing newline (UTF-8)")
	flag.BoolVar(emit, "t", true, "shortcut for -emit")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of decoder-ring:

    decoder-ring [-encode] <MODE>

MODE choices are %s, or an IANA encoding name. Modes marked with * are encode only.

`, getModes())
		flag.PrintDefaults()
	}
	flag.Parse()

	modeStr := flag.Arg(0)
	mode := modes[modeStr].decoder
	if *encode {
		mode = modes[modeStr].encoder
	}

	if mode == nil {
		i, err := ianaindex.IANA.Encoding(modeStr)
		if err == nil {
			if *encode {
				mode = i.NewEncoder().Bytes
			} else {
				mode = i.NewDecoder().Bytes
			}
		}
	}

	if flag.NArg() != 1 || mode == nil {
		flag.Usage()
		os.Exit(2)
	}

	if modeStr == "hex-extended" {
		*emit = false
	}

	if err := exec(mode, *strip, *emit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getModes() string {
	modesStr := make([]string, 0, len(modes))
	for mode, mf := range modes {
		if mf.decoder == nil {
			mode += "*"
		}
		modesStr = append(modesStr, mode)
	}
	sort.Strings(modesStr)
	return strings.Join(modesStr, ", ")
}

func exec(f modeFunc, stripNewline, emitNewline bool) error {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	if stripNewline {
		if len(b) > 0 && b[len(b)-1] == '\n' {
			b = b[:len(b)-1]
		}
	}
	b, err = f(b)
	if err != nil {
		return err
	}
	var trailer string
	if emitNewline {
		trailer = "\n"
	}
	_, err = io.Copy(os.Stdout, io.MultiReader(
		bytes.NewReader(b),
		strings.NewReader(trailer),
	))
	return err
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

func hexExtEnc(src []byte) (dst []byte, err error) {
	return []byte(hex.Dump(src)), nil
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

func rot13(src []byte) (dst []byte, err error) {
	dst = src[:0]
	for _, b := range src {
		if b >= 'A' && b <= 'Z' {
			n := (b - 'A' + 13) % 26
			b = 'A' + n
		} else if b >= 'a' && b <= 'z' {
			n := (b - 'a' + 13) % 26
			b = 'a' + n
		}
		dst = append(dst, b)
	}
	return
}

func base32Enc(src []byte) (dst []byte, err error) {
	dst = make([]byte, base32.StdEncoding.EncodedLen(len(src)))
	base32.StdEncoding.Encode(dst, src)
	return
}

func base32Dec(src []byte) ([]byte, error) {
	dst := make([]byte, base32.StdEncoding.DecodedLen(len(src)))
	n, err := base32.StdEncoding.Decode(dst, src)
	return dst[:n], err
}

func base32HexEnc(src []byte) (dst []byte, err error) {
	dst = make([]byte, base32.HexEncoding.EncodedLen(len(src)))
	base32.HexEncoding.Encode(dst, src)
	return
}

func base32HexDec(src []byte) ([]byte, error) {
	dst := make([]byte, base32.HexEncoding.DecodedLen(len(src)))
	n, err := base32.HexEncoding.Decode(dst, src)
	return dst[:n], err
}

var crockfordEnc = base32.NewEncoding("0123456789ABCDEFGHJKMNPQRSTVWXYZ")

func base32CrockfordEnc(src []byte) (dst []byte, err error) {
	dst = make([]byte, crockfordEnc.EncodedLen(len(src)))
	crockfordEnc.Encode(dst, src)
	return
}

func base32CrockfordDec(src []byte) ([]byte, error) {
	src = bytes.ToUpper(src)
	src = bytes.Replace(src, []byte("I"), []byte("1"), -1)
	src = bytes.Replace(src, []byte("L"), []byte("1"), -1)
	src = bytes.Replace(src, []byte("O"), []byte("0"), -1)
	src = bytes.Replace(src, []byte("-"), nil, -1)
	dst := make([]byte, crockfordEnc.DecodedLen(len(src)))
	n, err := crockfordEnc.Decode(dst, src)
	return dst[:n], err
}

func goEnc(src []byte) ([]byte, error) {
	return []byte(strconv.QuoteToASCII(string(src))), nil
}

func goDec(src []byte) ([]byte, error) {
	s := string(src)
	if len(src) > 0 && src[0] != '"' && src[0] != '`' && src[0] != '\'' {
		s = "\"" + s + "\""
	}
	s, err := strconv.Unquote(s)
	return []byte(s), err
}

func jsonEnc(src []byte) ([]byte, error) {
	s := string(src)
	b, err := json.Marshal(&s)
	return b, err
}

func jsonDec(src []byte) (dst []byte, err error) {
	if len(src) > 0 && src[0] != '"' {
		dst = make([]byte, len(src)+2)
		dst[0] = '"'
		dst[len(dst)-1] = '"'
		copy(dst[1:], src)
		src = dst
	}
	var s string
	err = json.Unmarshal(src, &s)
	dst = []byte(s)
	return
}

func htmlEnc(src []byte) ([]byte, error) {
	return []byte(html.EscapeString(string(src))), nil
}

func htmlDec(src []byte) ([]byte, error) {
	return []byte(html.UnescapeString(string(src))), nil
}

func urlPathEnc(src []byte) ([]byte, error) {
	return []byte(url.PathEscape(string(src))), nil
}

func urlPathDec(src []byte) ([]byte, error) {
	s, err := url.PathUnescape(string(src))
	return []byte(s), err
}

func urlQueryEnc(src []byte) ([]byte, error) {
	return []byte(url.QueryEscape(string(src))), nil
}

func urlQueryDec(src []byte) ([]byte, error) {
	s, err := url.QueryUnescape(string(src))
	return []byte(s), err
}
