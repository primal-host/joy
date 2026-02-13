package main

import "fmt"

type Parser struct {
	tokens  []Token
	pos     int
	machine *Machine
	scopes  []map[string]string // stack of {originalName → mangledName}
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

// pushScope increments the machine scope counter and pushes a new scope map.
func (p *Parser) pushScope() int {
	p.machine.ScopeID++
	id := p.machine.ScopeID
	p.scopes = append(p.scopes, map[string]string{})
	return id
}

// popScope removes the innermost scope map.
func (p *Parser) popScope() {
	if len(p.scopes) > 0 {
		p.scopes = p.scopes[:len(p.scopes)-1]
	}
}

// Parse processes all tokens, handling DEFINE and HIDE blocks and returning the remaining program.
func (p *Parser) Parse() []Value {
	var program []Value
	for !p.atEnd() {
		switch p.peek().Typ {
		case TokDefine:
			p.parseDefine()
		case TokHide:
			p.parseHide()
		default:
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
		if p.peek().Typ == TokHide {
			p.parseHide()
			continue
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

// parseHide handles HIDE ... IN ... END scoping.
func (p *Parser) parseHide() {
	p.advance() // consume HIDE
	scopeID := p.pushScope()
	prefix := fmt.Sprintf("__scope_%d_", scopeID)

	// Parse hidden definitions until IN
	p.parseDefSequence(prefix)

	// Expect IN
	if !p.atEnd() && p.peek().Typ == TokIn {
		p.advance() // consume IN
	} else {
		joyErr("expected IN after HIDE definitions")
	}

	// Parse public definitions until END (no mangling)
	p.parseDefSequence("")

	// Expect END
	if !p.atEnd() && p.peek().Typ == TokEnd {
		p.advance() // consume END
	} else {
		joyErr("expected END after IN definitions")
	}

	p.popScope()
}

// parseDefSequence parses a sequence of name == body definitions.
// If prefix is non-empty, names are mangled and registered in the current scope.
// Stops at IN, END, or end of input.
func (p *Parser) parseDefSequence(prefix string) {
	for !p.atEnd() {
		tok := p.peek()
		if tok.Typ == TokIn || tok.Typ == TokEnd || tok.Typ == TokDot {
			break
		}
		if tok.Typ == TokHide {
			p.parseHide()
			continue
		}
		if tok.Typ == TokDefine {
			p.advance() // consume DEFINE keyword (optional inside HIDE blocks)
			continue
		}
		if tok.Typ != TokAtom {
			joyErr("expected atom in definition, got %s", tok.Str)
		}
		name := p.advance().Str
		if p.peek().Typ != TokEqDef {
			joyErr("expected == after %s", name)
		}
		p.advance() // consume ==

		// Register in scope BEFORE parsing body (enables recursive self-references)
		dictName := name
		if prefix != "" {
			dictName = prefix + name
			p.scopes[len(p.scopes)-1][name] = dictName
		}

		body := p.readBody()
		p.machine.Dict[dictName] = body

		// consume optional ;
		if !p.atEnd() && p.peek().Typ == TokSemiCol {
			p.advance()
		}
	}
}

// readBody reads values until ; or . or IN or END or HIDE (at top level of DEFINE/HIDE)
func (p *Parser) readBody() []Value {
	var body []Value
	for !p.atEnd() {
		tok := p.peek()
		if tok.Typ == TokSemiCol || tok.Typ == TokDot || tok.Typ == TokIn || tok.Typ == TokEnd || tok.Typ == TokHide {
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
	// Check scope stack (inner to outer) for mangled name
	for i := len(p.scopes) - 1; i >= 0; i-- {
		if mangled, ok := p.scopes[i][name]; ok {
			return UserDefVal(mangled)
		}
	}
	return UserDefVal(name)
}
