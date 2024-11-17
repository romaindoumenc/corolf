package main

import (
	"unicode/utf8"
)

// Scanner implements a JSON Scanner as defined in RFC 7159.
type Scanner struct {
	offset int
}

const (
	exit = iota
	matchAny
	matchValue
	matchUnquoted
	matchQuoted
	// DoubleQuoted is not a valid output, but it is still found in some cases
	// This needs to be cleaned up in the log generator
	matchDoubleQuoted
	matchMessage
	toEOL
)

// scan reads ahead the next token, chosen depending on the current parser state.
// it sets the new state, the value and the offset
func (s *Parser) scan(next func(byte, []byte) int) {
	w := s.br.Window()
	curs := 0
	state := matchAny

	for {
		for pos, c := range w {
			if c > utf8.RuneSelf {
				continue
			}
			switch state {
			case exit:
				s.br.Release(curs)
				return

			case matchAny:
				switch c {
				case ' ', '\t', '\n':
					// skip leading whitespace
					curs = pos + 1
				case '\'':
					state = matchQuoted
					curs = pos + 1
				case '"':
					state = matchDoubleQuoted
					curs = pos + 1
				case '[':
					state = matchUnquoted
					curs = pos + 1
				default:
					state = matchUnquoted
				}

			case matchValue:
				switch c {
				case '\'':
					state = matchQuoted
					curs = pos + 1
				case '"':
					state = matchDoubleQuoted
					curs = pos + 1
				default:
					state = matchUnquoted
				}

			case matchQuoted:
				switch c {
				case '\'':
					c := byte(0)
					if pos < len(w)-1 {
						c = w[pos+1]
					}
					state = next(c, w[curs:pos])
					curs = pos + 1
				default:
				}

			case matchDoubleQuoted:
				switch c {
				case '"':
					c := byte(0)
					if pos < len(w)-1 {
						c = w[pos+1]
					}
					state = next(c, w[curs:pos])
					curs = pos + 1
				default:
				}

			case matchUnquoted:
				switch c {
				case '=', '[', ']', ' ':
					state = next(c, w[curs:pos])
					curs = pos + 1
				default:
				}

			case matchMessage:
				switch c {
				case ' ', '\t', ']':
					// skip leading whitespace and leftover symbol from double quoted string
					curs = pos + 1
				case '\n':
					// empty messages are dealt just like matches
					state = next(c, w[curs:pos])
					curs = pos + 1
				default:
					state = toEOL
				}

			case toEOL:
				switch c {
				case '\n':
					state = next(c, w[curs:pos])
					curs = pos + 1
				}
			}
		}

		// clean up already read data
		s.br.Release(curs)
		curs = 0

		// refill buffer
		if s.br.Extend() == 0 {
			// eof
			return
		}
		w = s.br.Window()
	}
}
