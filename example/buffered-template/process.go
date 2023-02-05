package impl

import (
	"fmt"
	"strings"

	"github.com/zalgonoise/parse"
)

// processFn is the ProcessFn that will process the Tree's Nodes, returning a string and an error
func processFn[C TextToken, T rune, R string](t *parse.Tree[C, T]) (R, error) {
	var sb = new(strings.Builder)
	for _, n := range t.List() {
		switch n.Type {
		case (C)(TokenIDENT):
			proc, err := processText[C, T, R](n)
			if err != nil {
				return (R)(sb.String()), err
			}
			sb.WriteString((string)(proc))
		case (C)(TokenLBRACE):
			proc, err := processTemplate[C, T, R](n)
			if err != nil {
				return (R)(sb.String()), err
			}
			sb.WriteString((string)(proc))
		}
	}

	return (R)(sb.String()), nil
}

// processText converts the T-type items into runes, and returns a string value of it
func processText[C TextToken, T rune, R string](n *parse.Node[C, T]) (R, error) {
	var val = make([]rune, len(n.Value), len(n.Value))
	for idx, r := range n.Value {
		val[idx] = (rune)(r)
	}
	return (R)(val), nil
}

// processTemplate prcesses the text within two template nodes
//
// Returns an error if a template is not terminated appropriately
func processTemplate[C TextToken, T rune, R string](n *parse.Node[C, T]) (R, error) {
	var sb = new(strings.Builder)
	var ended bool

	sb.WriteString(">>")
	for _, node := range n.Edges {
		switch node.Type {
		case (C)(TokenIDENT):
			proc, err := processText[C, T, R](node)
			if err != nil {
				return (R)(sb.String()), err
			}
			for _, e := range node.Edges {
				if e.Type == (C)(TokenRBRACE) {
					ended = true
				}
			}
			sb.WriteString((string)(proc))
		case (C)(TokenLBRACE):
			proc, err := processTemplate[C, T, R](node)
			if err != nil {
				return (R)(sb.String()), err
			}
			sb.WriteString((string)(proc))
		}
	}
	if !ended {
		return (R)(sb.String()), fmt.Errorf("parse error on line: %d", n.Pos)
	}

	sb.WriteString("<<")
	return (R)(sb.String()), nil
}
