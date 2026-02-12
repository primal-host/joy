package main

type Parser struct {
	tokens  []Token
	pos     int
	machine *Machine
}

func NewParser(tokens []Token, m *Machine) *Parser {
	return &Parser{tokens: tokens, pos: 0, machine: m}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Typ: TokEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Typ == TokEOF
}

// Parse processes all tokens, handling DEFINE blocks and returning the remaining program.
func (p *Parser) Parse() []Value {
	var program []Value
	for !p.atEnd() {
		if p.peek().Typ == TokDefine {
			p.parseDefine()
		} else {
			program = append(program, p.parseTerm()...)
		}
	}
	return program
}

func (p *Parser) parseDefine() {
	p.advance() // consume DEFINE
	for !p.atEnd() {
		if p.peek().Typ == TokDot {
			p.advance() // consume trailing .
			break
		}
		// expect: name == body ;|.
		if p.peek().Typ != TokAtom {
			joyErr("expected atom in DEFINE, got %s", p.peek().Str)
		}
		name := p.advance().Str
		if p.peek().Typ != TokEqDef {
			joyErr("expected == after %s in DEFINE", name)
		}
		p.advance() // consume ==
		body := p.readBody()
		p.machine.Dict[name] = body
		// consume optional ;
		if !p.atEnd() && p.peek().Typ == TokSemiCol {
			p.advance()
		}
	}
}

// readBody reads values until ; or . (at top level of DEFINE)
func (p *Parser) readBody() []Value {
	var body []Value
	for !p.atEnd() {
		tok := p.peek()
		if tok.Typ == TokSemiCol || tok.Typ == TokDot {
			break
		}
		// check if next atom looks like start of next definition (atom followed by ==)
		if tok.Typ == TokAtom && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Typ == TokEqDef {
			break
		}
		vals := p.parseTerm()
		body = append(body, vals...)
	}
	return body
}

// parseTerm reads one value (possibly a list) from the token stream.
func (p *Parser) parseTerm() []Value {
	tok := p.peek()

	switch tok.Typ {
	case TokInt:
		p.advance()
		return []Value{IntVal(tok.Int)}
	case TokFloat:
		p.advance()
		return []Value{FloatVal(tok.Flt)}
	case TokChar:
		p.advance()
		return []Value{CharVal(tok.Int)}
	case TokString:
		p.advance()
		return []Value{StringVal(tok.Str)}
	case TokLBrack:
		return []Value{p.parseList()}
	case TokLBrace:
		return []Value{p.parseSet()}
	case TokDot:
		p.advance()
		return []Value{p.resolveAtom(".")}
	case TokAtom:
		p.advance()
		return []Value{p.resolveAtom(tok.Str)}
	case TokEOF:
		return nil
	default:
		p.advance()
		joyErr("unexpected token: %s", tok.Str)
		return nil
	}
}

func (p *Parser) parseList() Value {
	p.advance() // consume [
	var items []Value
	for !p.atEnd() && p.peek().Typ != TokRBrack {
		items = append(items, p.parseTerm()...)
	}
	if !p.atEnd() {
		p.advance() // consume ]
	}
	if items == nil {
		items = []Value{}
	}
	return ListVal(items)
}

func (p *Parser) parseSet() Value {
	p.advance() // consume {
	var bits int64
	for !p.atEnd() && p.peek().Typ != TokRBrace {
		tok := p.advance()
		var n int64
		switch tok.Typ {
		case TokInt:
			n = tok.Int
		case TokChar:
			n = tok.Int
		default:
			joyErr("set members must be small integers or characters, got %s", tok.Str)
			continue
		}
		if n < 0 || n >= SetSize {
			joyErr("set member %d out of range 0..%d", n, SetSize-1)
		}
		bits |= 1 << n
	}
	if !p.atEnd() {
		p.advance() // consume }
	}
	return SetVal(bits)
}

func (p *Parser) resolveAtom(name string) Value {
	if fn, ok := builtins[name]; ok {
		return BuiltinVal(name, fn)
	}
	// Special literal atoms
	switch name {
	case "true":
		return BoolVal(true)
	case "false":
		return BoolVal(false)
	}
	return UserDefVal(name)
}
