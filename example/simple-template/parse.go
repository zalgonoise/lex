package impl

import (
	"github.com/zalgonoise/parse"
)

// initParse describes the ParseFn to kick off the parser. It is also the default fallback
// for any other ParseFn
func initParse[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != C(TokenEOF) {
		switch t.Peek().Type {
		case (C)(TokenIDENT):
			return parseText[C, T]
		case (C)(TokenLBRACE), (C)(TokenRBRACE):
			return parseTemplate[C, T]
		}
	}
	return nil
}

// parseText consumes the next item as a text token, creating a node for it under the
// current one in the tree.
func parseText[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	t.Node(t.Next())
	return initParse[C, T]
}

// parseTemplate creates a node for a template item, for which it expects both a text item edge
// that which also needs to contain an end-template edge.
//
// If it encounters a `}` token to close the template, it sets the position up three levels
// (back to the template's parent)
func parseTemplate[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	switch t.Peek().Type {
	case (C)(TokenLBRACE):
		t.Set(t.Parent())
		t.Node(t.Next())
	case (C)(TokenRBRACE):
		t.Node(t.Next())
		t.Set(t.Parent().Parent.Parent)
	}
	return initParse[C, T]
}
