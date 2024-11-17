package main

import (
	"fmt"
	"io"
	"strings"
	_ "unsafe"
)

type Parser struct{ br Reader }

func (p *Parser) Next() (string, error) {
	var tk byte
	var val []byte
	var st int

	var eof bool
	c := newcoro(func(c *coro) {
		p.scan(func(b1 byte, b2 []byte) int {
			tk, val = b1, b2
			coroswitch(c)
			return st
		})
		eof = true
	})

	for {
		rc := parse(func(i int) (byte, []byte) {
			st = i
			// throw garbage to parser until function terminates
			if eof {
				return 0, nil
			}

			coroswitch(c)
			return tk, val
		})

		if rc != "" {
			// terminate scanner
			if !eof {
				st = exit
				coroswitch(c)
			}

			return rc, nil
		}

		if eof {
			return "", io.EOF
		}

		// resync state: go to next EOL
		st = toEOL
		coroswitch(c)
	}
}

func parse(scan func(int) (byte, []byte)) (zero string) {
	var builder strings.Builder

	tk, val := scan(matchAny)
	if tk != ']' {
		return
	}

	fmt.Fprintf(&builder, "level=%s ", val)

	tk, val = scan(matchUnquoted)
	if tk != '[' {
		return
	}

	fmt.Fprintf(&builder, "timestamp=%s ", val)
	tk, val = scan(matchAny)
	if tk != ']' && tk != ' ' {
		return
	}

	fmt.Fprintf(&builder, "component=%s ", val)

	if tk == ' ' {
		for tk != ']' && tk != 0 {
			tk, val = scan(matchAny)
			if tk != '=' {
				return
			}

			key := string(val)
			tk, val = scan(matchValue)
			fmt.Fprintf(&builder, "attr.%s=%s ", key, val)
		}
	}

	tk, val = scan(matchMessage)
	if tk != '\n' {
		return
	}

	fmt.Fprintf(&builder, "message=%s", val)
	return builder.String()
}

// runtime tricks
//
// the usual parser / lexer can be a pain if know what too look up next.
//
// see https://github.com/golang/go/issues/67401

type coro struct{}

//go:linkname newcoro runtime.newcoro
func newcoro(func(*coro)) *coro

//go:linkname coroswitch runtime.coroswitch
func coroswitch(*coro)
