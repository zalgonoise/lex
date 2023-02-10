package protofile

import (
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/parse"
)

func initParse[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != C(TokenEOF) {
		switch t.Peek().Type {
		case (C)(TokenSYNTAX):
			return parseSyntax[C, T]
		case (C)(TokenPACKAGE):
			return parsePackage[C, T]
		case (C)(TokenENUM):
			return parseEnum[C, T]
		case (C)(TokenMESSAGE):
			return parseMessage[C, T]
		default:
			return nil
		}
	}
	return nil
}

// func parseText[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
// 	item := t.Next()
// 	if tok, ok := keywords[toString(item.Value)]; ok {
// 		item.Type = (C)(tok)

// 		t.Node(item)
// 		switch tok {
// 		case TokenSYNTAX:
// 			return parseSyntax[C, T]
// 		case TokenPACKAGE:
// 			return parsePackage[C, T]
// 		case TokenENUM:
// 			return parseEnum[C, T]
// 		case TokenMESSAGE:
// 			return parseMessage[C, T]
// 		default:
// 			return initParse[C, T]
// 		}
// 	}

//		t.Node(item)
//		return initParse[C, T]
//	}

func toString[T byte](v []T) string {
	buf := make([]byte, len(v))
	for i, b := range v {
		buf[i] = (byte)(b)
	}
	return (string)(buf)
}

func parseSyntax[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	t.Node(t.Next())
	if t.Peek().Type == C(TokenEQUAL) {
		t.Next()
	}
	if t.Peek().Type == C(TokenDQUOTE) {
		t.Next()
	}
	if t.Peek().Type == (C)(TokenIDENT) {
		item := t.Next()
		item.Type = C(TokenVALUE)
		t.Node(item)
	}
	if t.Peek().Type == C(TokenDQUOTE) {
		t.Next()
	}
	if t.Peek().Type == (C)(TokenSEMICOL) {
		t.Next() // skip
	}
	t.Set(t.Parent().Parent)
	return initParse[C, T]
}

func parsePackage[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	t.Node(t.Next())
	if t.Peek().Type == (C)(TokenIDENT) {
		item := t.Next()
		item.Type = C(TokenVALUE)
		t.Node(item)
		if t.Peek().Type == (C)(TokenSEMICOL) {
			t.Next() // skip
		}
	}
	t.Set(t.Parent().Parent)
	return initParse[C, T]
}

func parseEnum[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	t.Node(t.Next())
	if t.Peek().Type == (C)(TokenIDENT) {
		item := t.Next()
		item.Type = (C)(TokenTYPE)
		types[toString(item.Value)] = struct{}{}
		t.Node(item)
		if t.Peek().Type == (C)(TokenLBRACE) {
			t.Next() // skip
		}
		return parseKeyValue[C, T]
	}
	return initParse[C, T]
}

func parseKeyValue[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != (C)(TokenRBRACE) {

		var elem lex.Item[C, T]
		if t.Peek().Type == (C)(TokenIDENT) {
			elem = t.Next()
		}
		if t.Peek().Type == (C)(TokenEQUAL) {
			t.Next()
		}
		if t.Peek().Type == (C)(TokenIDENT) {
			item := t.Next()
			item.Type = (C)(TokenVALUE)
			t.Node(item)
			t.Node(elem)
		}
		if t.Peek().Type == (C)(TokenSEMICOL) {
			t.Next()
		}
		t.Set(t.Parent().Parent)
	}
	if t.Peek().Type == (C)(TokenRBRACE) {
		t.Next()
	}
	t.Set(t.Parent().Parent)
	return initParse[C, T]
}

func parseMessage[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	t.Node(t.Next())
	if t.Peek().Type == (C)(TokenIDENT) {
		item := t.Next()
		item.Type = (C)(TokenTYPE)
		types[toString(item.Value)] = struct{}{}
		t.Node(item)
		if t.Peek().Type == (C)(TokenLBRACE) {
			t.Next() // skip
		}
		return parseTypeKeyValue[C, T]
	}
	return initParse[C, T]
}

func parseTypeKeyValue[C ProtoToken, T byte](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != (C)(TokenRBRACE) {
		var elems = make([]lex.Item[C, T], 3)
		var repeated bool

		switch t.Peek().Type {
		case (C)(TokenREPEATED):
			elems[2] = t.Next()
			repeated = true
		case (C)(TokenMESSAGE):
			// t.Set(t.Parent())
			t.Store(parse.Slot0)
			return parseMessage[C, T]
		}

		if t.Peek().Type == (C)(TokenIDENT) {
			item := t.Next()
			_, ok := types[toString(item.Value)]
			if ok {
				item.Type = (C)(TokenTYPE)
				elems[0] = item
			}
		}
		if t.Peek().Type == (C)(TokenIDENT) {
			elems[1] = t.Next()
		}
		if t.Peek().Type == (C)(TokenEQUAL) {
			t.Next()
		}
		if t.Peek().Type == (C)(TokenIDENT) {
			item := t.Next()
			item.Type = (C)(TokenVALUE)
			t.Node(item)
			t.Node(elems[0])
			t.Set(t.Parent())
			t.Node(elems[1])
			t.Set(t.Parent())
			if repeated {
				t.Node(elems[2])
				t.Set(t.Parent())
			}
			t.Set(t.Parent())
		}
		if t.Peek().Type == (C)(TokenSEMICOL) {
			t.Next()
		}
	}
	if t.Peek().Type == (C)(TokenRBRACE) {
		t.Next()
	}
	if n := t.Load(parse.Slot0); n != nil {
		t.Set(n)
		return parseTypeKeyValue[C, T]
	}
	t.Set(t.Parent().Parent)
	return initParse[C, T]
}
