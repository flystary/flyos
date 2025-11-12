package dsl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}

func NewParser(input string) *Parser {
	l := NewLexer(input)
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Parse 解析 DSL，返回命令列表
func (p *Parser) Parse() ([]Command, error) {
	var cmds []Command
	for p.curToken.Type != TT_EOF {
		if cmd, ok := p.parseStatement(); ok && cmd != nil {
			cmds = append(cmds, *cmd)
		} else {
			p.nextToken()
		}
	}
	if len(p.errors) > 0 {
		return cmds, errors.New(strings.Join(p.errors, ", "))
	}
	return cmds, nil
}

// parseStatement 解析单条命令或 sync 块
func (p *Parser) parseStatement() (*Command, bool) {
	if p.curToken.Type != TT_IDENT {
		return nil, false
	}
	kind := strings.ToLower(p.curToken.Literal)

	// sync 块
	if strings.HasSuffix(kind, "s") && p.peekToken.Type == TT_SYNC {
		return p.parseSyncBlock(kind[:len(kind)-1])
	}

	// add/set/delete
	p.nextToken()
	verb := strings.ToLower(p.curToken.Literal)
	if verb != "add" && verb != "set" && verb != "delete" {
		p.error("expected verb add/set/delete")
		return nil, false
	}

	subtype := ""
	if p.peekToken.Type == TT_IDENT {
		p.nextToken()
		subtype = p.curToken.Literal
	}

	p.expect(TT_LBRACE)
	attrs := p.parseAttributes()
	p.expect(TT_RBRACE)

	return &Command{
		Kind:    kind,
		Verb:    verb,
		Subtype: subtype,
		Attrs:   attrs,
	}, true
}

// parseSyncBlock 解析 sync 块
func (p *Parser) parseSyncBlock(kind string) (*Command, bool) {
	p.expect(TT_SYNC)   // consume SYNC
	p.expect(TT_LBRACE) // consume {

	var blocks []Command
	for p.peekToken.Type != TT_RBRACE && p.peekToken.Type != TT_EOF {
		p.nextToken()
		subtype := p.curToken.Literal

		p.expect(TT_LBRACE)
		attrs := p.parseAttributes()
		p.expect(TT_RBRACE)

		blocks = append(blocks, Command{
			Kind:    kind,
			Verb:    "sync",
			Subtype: subtype,
			Attrs:   attrs,
		})
	}

	p.expect(TT_RBRACE)
	return &Command{
		Kind:   kind,
		Verb:   "sync",
		Blocks: blocks,
	}, true
}

// parseAttributes 解析 key/value 属性
func (p *Parser) parseAttributes() map[string]interface{} {
	attrs := map[string]interface{}{}
	for p.peekToken.Type != TT_RBRACE && p.peekToken.Type != TT_EOF {
		p.nextToken()
		if p.curToken.Type != TT_IDENT {
			p.error("expected attribute key")
			continue
		}
		key := p.curToken.Literal

		p.nextToken()
		var val interface{}
		switch p.curToken.Type {
		case TT_STRING:
			val = p.curToken.Literal
		case TT_NUMBER:
			if i, err := strconv.Atoi(p.curToken.Literal); err == nil {
				val = i
			} else {
				val = p.curToken.Literal
			}
		case TT_BOOL:
			v := strings.ToLower(p.curToken.Literal)
			val = (v == "yes" || v == "true")
		case TT_LBRACK:
			var items []string
			for {
				p.nextToken()
				if p.curToken.Type == TT_RBRACK || p.curToken.Type == TT_EOF {
					break
				}
				if p.curToken.Type == TT_IDENT || p.curToken.Type == TT_STRING || p.curToken.Type == TT_NUMBER {
					items = append(items, p.curToken.Literal)
				}
				if p.peekToken.Type == TT_COMMA {
					p.nextToken()
				}
			}
			val = items
		default:
			val = p.curToken.Literal
		}

		attrs[key] = val
		if p.peekToken.Type == TT_SEMI {
			p.nextToken()
		}
	}
	return attrs
}

func (p *Parser) expect(t TokenType) {
	if p.peekToken.Type != t {
		p.error(fmt.Sprintf("expected %s, got %s", t, p.peekToken.Type))
	}
	p.nextToken()
}

func (p *Parser) error(msg string) {
	p.errors = append(p.errors, msg)
}
