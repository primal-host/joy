package main

import (
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokAtom    TokenType = iota // identifier or operator name
	TokInt                      // integer literal
	TokFloat                    // float literal
	TokChar                     // character literal 'x
	TokString                   // string literal "..."
	TokLBrack                   // [
	TokRBrack                   // ]
	TokLBrace                   // {
	TokRBrace                   // }
	TokDot                      // .  (auto-put)
	TokSemiCol                  // ;
	TokDefine                   // DEFINE keyword
	TokEqDef                    // ==
	TokEOF
)

type Token struct {
	Typ TokenType
	Str string  // raw text for atoms, string value for strings
	Int int64   // integer or char value
	Flt float64 // float value
}

type Scanner struct {
	src []rune
	pos int
}

func NewScanner(source string) *Scanner {
	return &Scanner{src: []rune(source), pos: 0}
}

func (s *Scanner) atEnd() bool {
	return s.pos >= len(s.src)
}

func (s *Scanner) peek() rune {
	if s.atEnd() {
		return 0
	}
	return s.src[s.pos]
}

func (s *Scanner) advance() rune {
	ch := s.src[s.pos]
	s.pos++
	return ch
}

func (s *Scanner) skipWhitespaceAndComments() {
	for !s.atEnd() {
		ch := s.peek()
		// whitespace
		if unicode.IsSpace(ch) {
			s.advance()
			continue
		}
		// line comment: # to end of line
		if ch == '#' {
			for !s.atEnd() && s.peek() != '\n' {
				s.advance()
			}
			continue
		}
		// block comment: (* ... *)
		if ch == '(' && s.pos+1 < len(s.src) && s.src[s.pos+1] == '*' {
			s.advance() // (
			s.advance() // *
			for !s.atEnd() {
				if s.peek() == '*' {
					s.advance()
					if !s.atEnd() && s.peek() == ')' {
						s.advance()
						break
					}
				} else {
					s.advance()
				}
			}
			continue
		}
		break
	}
}

func (s *Scanner) specialChar() rune {
	if s.atEnd() {
		return '\\'
	}
	ch := s.advance()
	switch ch {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'b':
		return '\b'
	case 'r':
		return '\r'
	case 'f':
		return '\f'
	case '\'':
		return '\''
	case '"':
		return '"'
	case '\\':
		return '\\'
	default:
		if ch >= '0' && ch <= '9' {
			num := int(ch - '0')
			for i := 0; i < 2 && !s.atEnd(); i++ {
				c := s.peek()
				if c >= '0' && c <= '9' {
					num = num*10 + int(c-'0')
					s.advance()
				} else {
					break
				}
			}
			return rune(num)
		}
		return ch
	}
}

func isAtomChar(ch rune) bool {
	if unicode.IsSpace(ch) {
		return false
	}
	switch ch {
	case '[', ']', '{', '}', '(', ')', ';', '"', '\'', '#':
		return false
	}
	return true
}

func (s *Scanner) ScanAll() []Token {
	var tokens []Token
	for {
		tok := s.Next()
		tokens = append(tokens, tok)
		if tok.Typ == TokEOF {
			break
		}
	}
	return tokens
}

func (s *Scanner) Next() Token {
	s.skipWhitespaceAndComments()
	if s.atEnd() {
		return Token{Typ: TokEOF}
	}

	ch := s.peek()

	switch ch {
	case '[':
		s.advance()
		return Token{Typ: TokLBrack, Str: "["}
	case ']':
		s.advance()
		return Token{Typ: TokRBrack, Str: "]"}
	case '{':
		s.advance()
		return Token{Typ: TokLBrace, Str: "{"}
	case '}':
		s.advance()
		return Token{Typ: TokRBrace, Str: "}"}
	case '.':
		// check if followed by digit (could be float like .5) — Joy doesn't use this, treat as period
		s.advance()
		return Token{Typ: TokDot, Str: "."}
	case ';':
		s.advance()
		return Token{Typ: TokSemiCol, Str: ";"}
	case '\'':
		s.advance()
		if s.atEnd() {
			joyErr("unexpected end of input after '")
		}
		var c rune
		if s.peek() == '\\' {
			s.advance()
			c = s.specialChar()
		} else {
			c = s.advance()
		}
		return Token{Typ: TokChar, Int: int64(c)}
	case '"':
		s.advance()
		var buf strings.Builder
		for !s.atEnd() && s.peek() != '"' {
			if s.peek() == '\\' {
				s.advance()
				buf.WriteRune(s.specialChar())
			} else {
				buf.WriteRune(s.advance())
			}
		}
		if !s.atEnd() {
			s.advance() // closing "
		}
		return Token{Typ: TokString, Str: buf.String()}
	}

	// numbers: optional leading minus followed by digits
	if ch >= '0' && ch <= '9' || (ch == '-' && s.pos+1 < len(s.src) && s.src[s.pos+1] >= '0' && s.src[s.pos+1] <= '9') {
		return s.scanNumber()
	}

	// atom (identifier or operator)
	return s.scanAtom()
}

func (s *Scanner) scanNumber() Token {
	start := s.pos
	if s.peek() == '-' {
		s.advance()
	}
	for !s.atEnd() && s.peek() >= '0' && s.peek() <= '9' {
		s.advance()
	}
	isFloat := false
	if !s.atEnd() && s.peek() == '.' && s.pos+1 < len(s.src) && s.src[s.pos+1] >= '0' && s.src[s.pos+1] <= '9' {
		isFloat = true
		s.advance() // .
		for !s.atEnd() && s.peek() >= '0' && s.peek() <= '9' {
			s.advance()
		}
	}
	if !s.atEnd() && (s.peek() == 'e' || s.peek() == 'E') {
		isFloat = true
		s.advance()
		if !s.atEnd() && (s.peek() == '+' || s.peek() == '-') {
			s.advance()
		}
		for !s.atEnd() && s.peek() >= '0' && s.peek() <= '9' {
			s.advance()
		}
	}
	text := string(s.src[start:s.pos])
	if isFloat {
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			joyErr("invalid float: %s", text)
		}
		return Token{Typ: TokFloat, Flt: f, Str: text}
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		joyErr("invalid integer: %s", text)
	}
	return Token{Typ: TokInt, Int: n, Str: text}
}

func (s *Scanner) scanAtom() Token {
	start := s.pos
	for !s.atEnd() && isAtomChar(s.peek()) {
		s.advance()
	}
	text := string(s.src[start:s.pos])
	if text == "" {
		ch := s.advance()
		joyErr("unexpected character: %c", ch)
	}
	switch text {
	case "DEFINE", "PUBLIC", "LIBRA":
		return Token{Typ: TokDefine, Str: text}
	case "==":
		return Token{Typ: TokEqDef, Str: text}
	default:
		return Token{Typ: TokAtom, Str: text}
	}
}
