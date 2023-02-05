package impl

import "github.com/zalgonoise/lex"

// initState describes the StateFn to kick off the lexer. It is also the default fallback StateFn
// for any other StateFn
func initState[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	switch l.Next() {
	case '}':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		l.Ignore()
		return stateRBRACE[C, T]
	case '{':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		l.Ignore()
		return stateLBRACE[C, T]
	case 0:
		return nil
	default:
		return stateIDENT[C, T]
	}
}

// stateIDENT describes the StateFn to parse text tokens.
func stateIDENT[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.AcceptRun(func(item T) bool {
		return item != '}' && item != '{' && item != 0
	})
	switch l.Next() {
	case '}':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		return stateRBRACE[C, T]
	case '{':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		return stateLBRACE[C, T]
	default:
		if l.Width() > 0 {
			l.Emit((C)(TokenIDENT))
		}
		l.Emit((C)(TokenEOF))
		return nil
	}
}

// stateLBRACE describes the StateFn to check for and emit an LBRACE token
func stateLBRACE[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.Next() // skip this symbol
	l.Emit((C)(TokenLBRACE))
	return initState[C, T]
}

// stateRBRACE describes the StateFn to check for and emit an RBRACE token
func stateRBRACE[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.Next() // skip this symbol
	l.Emit((C)(TokenRBRACE))
	return initState[C, T]
}

// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.Backup()
	l.Prev() // mark the previous char as erroring token
	l.Emit((C)(TokenError))
	return initState[C, T]
}
