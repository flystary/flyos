package dsl

import (
	"strconv"
	"strings"
	"unicode"
)

type TokenType string

const (
	TT_IDENT  TokenType = "IDENT"
	TT_NUMBER TokenType = "NUMBER"
	TT_STRING TokenType = "STRING"
	TT_BOOL   TokenType = "BOOL"
	TT_LBRACE TokenType = "{"
	TT_RBRACE TokenType = "}"
	TT_SEMI   TokenType = ";"
	TT_LBRACK TokenType = "["
	TT_RBRACK TokenType = "]"
	TT_COMMA  TokenType = ","
	TT_EOF    TokenType = "EOF"
	TT_SYNC   TokenType = "SYNC"
	TT_ADD    TokenType = "ADD"
	TT_SET    TokenType = "SET"
	TT_DELETE TokenType = "DELETE"
)

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readPos])
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.readPos])
}

func (l *Lexer) skipSpaceAndComments() {
	for {
		if unicode.IsSpace(l.ch) {
			l.readChar()
			continue
		}
		if l.ch == '#' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}
		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}
		break
	}
}

func isIdentChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == '.' || ch == ':'
}

func (l *Lexer) readIdent() string {
	start := l.pos
	for isIdentChar(l.ch) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readString() string {
	l.readChar()
	start := l.pos
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	val := l.input[start:l.pos]
	l.readChar()
	return val
}

func (l *Lexer) readNumberLike() string {
	start := l.pos
	for unicode.IsDigit(l.ch) || l.ch == '.' || l.ch == '/' || l.ch == ':' {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) NextToken() Token {
	l.skipSpaceAndComments()
	tok := Token{Pos: l.pos}
	switch l.ch {
	case '{':
		tok.Type = TT_LBRACE
		tok.Literal = "{"
	case '}':
		tok.Type = TT_RBRACE
		tok.Literal = "}"
	case ';':
		tok.Type = TT_SEMI
		tok.Literal = ";"
	case '[':
		tok.Type = TT_LBRACK
		tok.Literal = "["
	case ']':
		tok.Type = TT_RBRACK
		tok.Literal = "]"
	case ',':
		tok.Type = TT_COMMA
		tok.Literal = ","
	case '"':
		tok.Type = TT_STRING
		tok.Literal = l.readString()
		return tok
	case 0:
		tok.Type = TT_EOF
		return tok
	default:
		if unicode.IsLetter(l.ch) {
			ident := l.readIdent()
			lower := strings.ToLower(ident)
			switch lower {
			case "add":
				tok.Type = TT_ADD
			case "set":
				tok.Type = TT_SET
			case "delete":
				tok.Type = TT_DELETE
			case "sync":
				tok.Type = TT_SYNC
			case "yes", "no", "true", "false":
				tok.Type = TT_BOOL
			default:
				tok.Type = TT_IDENT
			}
			tok.Literal = ident
			return tok
		}
		if unicode.IsDigit(l.ch) {
			num := l.readNumberLike()
			if _, err := strconv.Atoi(num); err == nil {
				tok.Type = TT_NUMBER
			} else {
				tok.Type = TT_IDENT
			}
			tok.Literal = num
			return tok
		}
		tok.Type = TT_IDENT
		tok.Literal = string(l.ch)
	}
	l.readChar()
	return tok
}
