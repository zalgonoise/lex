package protofile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/parse"
)

func processFn[C ProtoToken, T byte, R string](t *parse.Tree[C, T]) (R, error) {
	var sb = new(strings.Builder)
	sb.WriteByte('\n')

	for _, n := range t.List() {

		switch n.Type {
		case C(TokenSYNTAX):
			processSyntax(sb, n)
		case C(TokenPACKAGE):
			processPackage(sb, n)
		case C(TokenENUM):
			processEnum(sb, n)
		case C(TokenMESSAGE):
			processMessage(sb, n, 0)
		default:
			return (R)(sb.String()), fmt.Errorf("invalid top-level token: %d -- %s", n.Type, toString(n.Value))
		}
	}

	return (R)(sb.String()), nil
}

func processSyntax[C ProtoToken, T byte](sb *strings.Builder, n *parse.Node[C, T]) {
	sb.WriteString("syntax: ")
	for _, e := range n.Edges {
		if e.Type == C(TokenVALUE) {
			sb.WriteString(toString(e.Value))
		}
	}
	sb.WriteByte('\n')
}

func processPackage[C ProtoToken, T byte](sb *strings.Builder, n *parse.Node[C, T]) {
	sb.WriteString("package: ")
	for _, e := range n.Edges {
		if e.Type == C(TokenVALUE) {
			sb.WriteString(toString(e.Value))
		}
	}
	sb.WriteByte('\n')
}

func processEnum[C ProtoToken, T byte](sb *strings.Builder, n *parse.Node[C, T]) {
	sb.WriteString("enum (len: ")
	if len(n.Edges) > 0 {
		sb.WriteString(strconv.Itoa(len(n.Edges[0].Edges)))
	}
	sb.WriteString(")\n")
	for _, e := range n.Edges {
		if e.Type == (C)(TokenTYPE) {
			sb.WriteString("type: ")
			sb.WriteString(toString(e.Value))
			sb.WriteByte('\n')
			for _, ee := range e.Edges {
				processEnumFields(sb, ee)
			}
		}
	}
}
func processEnumFields[C ProtoToken, T byte](sb *strings.Builder, n *parse.Node[C, T]) {
	if n.Type == (C)(TokenVALUE) {
		sb.WriteString("\tid: ")
		sb.WriteString(toString(n.Value))
		for _, e := range n.Edges {
			sb.WriteString("\tname: ")
			sb.WriteString(toString(e.Value))
		}
		sb.WriteByte('\n')
	}
}

func processMessage[C ProtoToken, T byte](sb *strings.Builder, n *parse.Node[C, T], ident int) {
	addIdent(sb, ident)
	sb.WriteString("message (len: ")
	if len(n.Edges) > 0 {
		sb.WriteString(strconv.Itoa(len(n.Edges[0].Edges)))
	}
	sb.WriteString(")\n")

	for _, e := range n.Edges {
		if e.Type == (C)(TokenTYPE) {
			addIdent(sb, ident)
			sb.WriteString("type: ")
			sb.WriteString(toString(e.Value))
			sb.WriteByte('\n')

			for _, ee := range e.Edges {
				processMessageFields(sb, ee, ident)
			}

		}
	}
}
func addIdent(sb *strings.Builder, n int) {
	for i := 0; i < n; i++ {
		sb.WriteByte('\t')
	}
}

func processMessageFields[C ProtoToken, T byte](sb *strings.Builder, n *parse.Node[C, T], ident int) {
	switch n.Type {
	case (C)(TokenVALUE):
		addIdent(sb, ident)
		sb.WriteString("\tid: ")
		sb.WriteString(toString(n.Value))
		for _, e := range n.Edges {
			switch e.Type {
			case (C)(TokenMESSAGE):
				processMessage(sb, e, ident+1)
			case C(TokenIDENT):
				sb.WriteString("\tname: ")
				sb.WriteString(toString(e.Value))
			case C(TokenTYPE):
				sb.WriteString("\ttype: ")
				sb.WriteString(toString(e.Value))
			case C(TokenREPEATED):
				sb.WriteString("\trepeated: true")
			}
		}
		sb.WriteByte('\n')
	case (C)(TokenMESSAGE):
		processMessage(sb, n, ident+1)
	}

}

func Run[C ProtoToken, T byte, R string](r gio.Reader[T]) (R, error) {
	var rootEOF C
	l := (lex.Emitter[C, T])(lex.NewBuffer(initState[C, T], r))
	t := parse.New(l, initParse[C, T], rootEOF)

	// for t.Peek().Type != C(TokenEOF) {
	// 	item := t.Next()
	// 	fmt.Println(item.Type, toString(item.Value))
	// }
	// return "", nil

	t.Parse()
	return processFn[C, T, R](t)
}
