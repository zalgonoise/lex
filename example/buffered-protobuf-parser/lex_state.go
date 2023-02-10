package protofile

import (
	"bytes"

	"github.com/zalgonoise/lex"
)

var chars = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_")

func initState[C ProtoToken, T byte](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	switch l.Next() {
	case '=':
		l.Emit((C)(TokenEQUAL))
		return initState[C, T]
	case '"':
		l.Emit((C)(TokenDQUOTE))
		return initState[C, T]
	case ';':
		l.Emit((C)(TokenSEMICOL))
		return initState[C, T]
	case '{':
		l.Emit((C)(TokenLBRACE))
		return initState[C, T]
	case '}':
		l.Emit((C)(TokenRBRACE))
		return initState[C, T]
	case ' ', '\t', '\n':
		l.Ignore()
		return initState[C, T]
	case 0:
		return nil
	default:
		return stateIDENT[C, T]
	}
}

func stateIDENT[C ProtoToken, T byte](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.AcceptRun(func(item T) bool {
		return bytes.IndexByte(chars, byte(item)) >= 0
	})
	if l.Width() > 0 {
		buf := l.Extract(l.Start(), l.Pos())

		tok, ok := keywords[toString(buf)]
		if ok {
			l.Emit((C)(tok))
			return initState[C, T]
		}
		l.Emit((C)(TokenIDENT))
	}

	return initState[C, T]
}
