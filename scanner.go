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
	TokHide                     // HIDE, PRIVATE
	TokIn                       // IN
	TokEnd                      // END
	TokModule                   // MODULE
	TokEqDef                    // ==
	TokEOF
)

type Token struct {
	Typ TokenType
	Str string  // raw text for atoms, string value for strings
	Int int64   // integer or char value
	Flt float64 // float value
	Col int     // 1-indexed column in source (0 = unknown)
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
		return Token{Typ: TokEOF, Col: s.pos + 1}
	}

	col := s.pos + 1 // 1-indexed column
	ch := s.peek()

	switch ch {
	case '[':
		s.advance()
		return Token{Typ: TokLBrack, Str: "[", Col: col}
	case ']':
		s.advance()
		return Token{Typ: TokRBrack, Str: "]", Col: col}
	case '{':
		s.advance()
		return Token{Typ: TokLBrace, Str: "{", Col: col}
	case '}':
		s.advance()
		return Token{Typ: TokRBrace, Str: "}", Col: col}
	case '.':
		s.advance()
		// .s and similar: dot followed by letter is an atom
		if !s.atEnd() && isAtomChar(s.peek()) && s.peek() != '.' {
			start := s.pos
			for !s.atEnd() && isAtomChar(s.peek()) {
				s.advance()
			}
			name := "." + string(s.src[start:s.pos])
			return Token{Typ: TokAtom, Str: name, Col: col}
		}
		return Token{Typ: TokDot, Str: ".", Col: col}
	case ';':
		s.advance()
		return Token{Typ: TokSemiCol, Str: ";", Col: col}
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
		return Token{Typ: TokChar, Int: int64(c), Col: col}
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
		return Token{Typ: TokString, Str: buf.String(), Col: col}
	}

	// numbers: optional leading minus followed by digits
	if ch >= '0' && ch <= '9' || (ch == '-' && s.pos+1 < len(s.src) && s.src[s.pos+1] >= '0' && s.src[s.pos+1] <= '9') {
		return s.scanNumber()
	}

	// atom (identifier or operator)
	return s.scanAtom()
}

func (s *Scanner) scanNumber() Token {
	col := s.pos + 1
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
		return Token{Typ: TokFloat, Flt: f, Str: text, Col: col}
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		joyErr("invalid integer: %s", text)
	}
	return Token{Typ: TokInt, Int: n, Str: text, Col: col}
}

func (s *Scanner) scanAtom() Token {
	col := s.pos + 1
	start := s.pos
	for !s.atEnd() && isAtomChar(s.peek()) {
		// Only include '.' if followed by another atom char (module dot-notation: m1.ab)
		// Don't include trailing '.' (statement terminator: foo.)
		if s.peek() == '.' {
			if s.pos+1 >= len(s.src) || !isAtomChar(rune(s.src[s.pos+1])) || s.src[s.pos+1] == '.' {
				break
			}
		}
		s.advance()
	}
	text := string(s.src[start:s.pos])
	if text == "" {
		ch := s.advance()
		joyErr("unexpected character: %c", ch)
	}
	switch text {
	case "DEFINE", "PUBLIC", "LIBRA":
		return Token{Typ: TokDefine, Str: text, Col: col}
	case "HIDE", "PRIVATE":
		return Token{Typ: TokHide, Str: text, Col: col}
	case "IN":
		return Token{Typ: TokIn, Str: text, Col: col}
	case "END":
		return Token{Typ: TokEnd, Str: text, Col: col}
	case "MODULE":
		return Token{Typ: TokModule, Str: text, Col: col}
	case "==":
		return Token{Typ: TokEqDef, Str: text, Col: col}
	default:
		return Token{Typ: TokAtom, Str: text, Col: col}
	}
}
