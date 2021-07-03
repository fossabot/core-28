package lexer

import (
	"errors"
	"fmt"

	"asciigoat.org/core/runes"
)

// state function
type StateFn func(Lexer) StateFn

type Lexer interface {
	Run() // run state machine

	Position() TokenPosition // base for the next token
	Tokens() <-chan Token    // tokens output

	AtLeast(n int) ([]rune, error)

	NewLine()
	Step(n int)

	Emit(TokenType)
	EmitError(error)
	EmitErrorf(string, ...interface{})
	EmitSyntaxError(string, ...interface{})
}

type lexer struct {
	start StateFn // initial state

	in     *runes.Feeder // runes source
	pos    TokenPosition // base for the next token
	cursor int           // look ahead pointer
	tokens chan Token    // tokens output
}

func NewLexer(start StateFn, in *runes.Feeder, tokens int) Lexer {
	return &lexer{
		start:  start,
		in:     in,
		pos:    TokenPosition{1, 1},
		tokens: make(chan Token, tokens),
	}
}

func (lex *lexer) Run() {
	defer close(lex.tokens)

	for state := lex.start; state != nil; {
		state = state(lex)
	}
}

func (lex *lexer) AtLeast(n int) ([]rune, error) {
	min := lex.cursor
	if n > 0 {
		min += n
	}

	s, err := lex.in.AtLeast(min)
	if len(s) > lex.cursor {
		s = s[lex.cursor:]
	} else {
		s = nil
	}
	return s, err
}

func (lex *lexer) Position() TokenPosition {
	return lex.pos
}

func (lex *lexer) Step(n int) {
	lex.cursor += n
}

func (lex *lexer) NewLine() {
	lex.pos.NewLine()
}

func (lex *lexer) Tokens() <-chan Token {
	return lex.tokens
}

func (lex *lexer) Emit(typ TokenType) {
	var text []rune

	pos := lex.pos

	// extract text to emit, and update cursor for the next
	if n := lex.cursor; n > 0 {
		text = lex.in.Runes()[:n]
		lex.in.Skip(n)
		lex.pos.Step(n)
		lex.cursor = 0
	}

	lex.tokens <- NewToken(typ, text, pos)
}

func (lex *lexer) EmitError(err error) {
	lex.tokens <- NewErrorToken(err, lex.pos)
}

func (lex *lexer) EmitErrorf(s string, args ...interface{}) {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	lex.tokens <- NewErrorToken(errors.New(s), lex.pos)
}

func (lex *lexer) EmitSyntaxError(s string, args ...interface{}) {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	lex.tokens <- NewSyntaxErrorToken(s, lex.pos, lex.cursor, lex.in.Runes())
}
