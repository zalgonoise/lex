package impl

import (
	"github.com/zalgonoise/lex"
)

// TextToken is a unique identifier for this text template implementation
type TextToken int

const (
	TokenEOF TextToken = iota
	TokenError
	TokenIDENT
	TokenLBRACE
	TokenRBRACE
)

// TemplateItem represents the lex.Item for a runes lexer based on TextToken identifiers
type TemplateItem[C TextToken, I rune] lex.Item[C, I]

// toTemplateItem converts a lex.Item type to TemplateItem
func toTemplateItem[C TextToken, T rune](i lex.Item[C, T]) TemplateItem[C, T] {
	return (TemplateItem[C, T])(i)
}
