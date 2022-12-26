package impl

import (
	"fmt"
	"strings"

	"github.com/zalgonoise/lex"
)

// TextToken is a unique identifier for this text template implementation
type TextToken int

const (
	TokenEOF TextToken = iota
	TokenError
	TokenIDENT
	TokenTEMPL
	TokenLBRACE
	TokenRBRACE
)

// TemplateItem represents the lex.Item for a runes lexer based on TextToken identifiers
type TemplateItem[C TextToken, I rune] lex.Item[C, I]

// toTemplateItem converts a lex.Item type to TemplateItem
func toTemplateItem[C TextToken, T rune](i lex.Item[C, T]) TemplateItem[C, T] {
	return (TemplateItem[C, T])(i)
}

// String implements fmt.Stringer; which is processing each TemplateItem as a string
func (t TemplateItem[C, I]) String() string {
	switch t.Type {
	case C(TokenIDENT):
		var rs = make([]rune, len(t.Value), len(t.Value))
		for idx, r := range t.Value {
			rs[idx] = (rune)(r)
		}
		return string(rs)
	case C(TokenTEMPL):
		var rs = make([]rune, len(t.Value), len(t.Value))
		for idx, r := range t.Value {
			rs[idx] = (rune)(r)
		}
		return ">>" + string(rs) + "<<"
	case C(TokenError):
		return ":ERR:"
	case C(TokenEOF):
		return "" // placeholder action for EOF tokens
	}
	return ""
}

// initState describes the StateFn to kick off the lexer. It is also the default fallback StateFn
// for any other StateFn
func initState[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.AcceptRun(func(item T) bool {
		return item != '{' && item != 0
	})

	switch l.Next() {
	case '{':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		l.Ignore()
		return stateLBRACE[C, T, I]
	default:
		if l.Width() > 0 {
			l.Emit((C)(TokenIDENT))
		}
		l.Emit((C)(TokenEOF))
		return nil
	}
}

// stateLBRACE describes the StateFn to read the template content, emitting it as a template item
func stateLBRACE[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	if l.Check(func(item T) bool {
		return item == '{'
	}) {
		l.Next() // skip this symbol
		l.Ignore()
	}

	l.AcceptRun(func(item T) bool {
		return item != '}' && item != 0
	})

	switch l.Next() {
	case '}':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenTEMPL))
		}
		l.Next() // skip this symbol
		l.Ignore()
		return initState[C, T, I]
	default:
		return stateError[C, T, I]
	}
}

// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.Emit((C)(TokenError))
	return initState[C, T, I]
}

// TextTemplateLexer creates a text template lexer based on the input slice of runes
func TextTemplateLexer[C TextToken, T rune, I lex.Item[C, T]](input []T) lex.Lexer[C, T, I] {
	return lex.New(initState[C, T, I], input)
}

// Run takes in a string `s`, processes it for templates, and returns the processed string and an error
func Run(s string) (string, error) {
	l := TextTemplateLexer([]rune(s))
	var sb = new(strings.Builder)
	for {
		i := l.NextItem()
		sb.WriteString(toTemplateItem(i).String())

		switch i.Type {
		case 0:
			return sb.String(), nil
		case TokenError:
			return sb.String(), fmt.Errorf("failed to parse token")
		}
	}
}
