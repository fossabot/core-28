package lexer

import (
	"errors"
	"fmt"
	"io"
)

var (
	EOF = io.EOF // EOF marker
)

// Token type
type TokenType int

const (
	TokenError TokenType = iota
)

// Token Position
type TokenPosition struct {
	Line int
	Row  int
}

func (pos *TokenPosition) Reset() {
	pos.Line = 1
	pos.Row = 1
}

func (pos *TokenPosition) Step(n int) {
	pos.Row += n
}

func (pos *TokenPosition) NewLine() {
	pos.Line += 1
	pos.Row = 1
}

// Token
type Token interface {
	Type() TokenType
	String() string
	Position() TokenPosition
}

type token struct {
	typ TokenType
	pos TokenPosition
	val string
}

func NewToken(typ TokenType, val []rune, pos TokenPosition) Token {
	return &token{
		typ: typ,
		val: string(val),
		pos: pos,
	}
}

func (t token) Type() TokenType {
	return t.typ
}

func (t token) Position() TokenPosition {
	return t.pos
}

func (t token) String() string {
	return t.val
}

// ErrorToken
type ErrorToken interface {
	Token
	Error() string
	Unwrap() error
}

type errorToken struct {
	token
	err error
}

func NewErrorToken(err error, pos TokenPosition) ErrorToken {
	return &errorToken{
		token: token{
			typ: TokenError,
			val: err.Error(),
			pos: pos,
		},
		err: err,
	}
}

func (t errorToken) Error() string {
	return t.err.Error()
}

func (t errorToken) Unwrap() error {
	return t.err
}

// SyntaxErrorToken
type SyntaxErrorToken struct {
	ErrorToken

	Cursor int
	Buffer string
}

func NewSyntaxErrorToken(msg string, pos TokenPosition, cur int, buffer []rune) *SyntaxErrorToken {
	s := fmt.Sprintf("Syntax Error at %v.%v+%v", pos.Line, pos.Row, cur)

	if len(msg) > 0 {
		s = fmt.Sprintf("%s: %s", s, msg)
	}

	return &SyntaxErrorToken{
		ErrorToken: NewErrorToken(errors.New(s), pos),

		Cursor: cur,
		Buffer: string(buffer),
	}
}
