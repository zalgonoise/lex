package lex_test

import (
	"strings"
	"testing"

	"github.com/zalgonoise/lex"
)

const (
	tokenEOF uint = iota
	tokenError
	tokenIdent
	tokenPeriod
)

var (
	testInput1 = []rune{'l', 'e', 'x', 'i', 'n', 'g', ' ', 'd', 'a', 't', 'a', '.'}
	testInput2 = []rune{'l', 'e', 'x', 'i', 'n', 'g', '.', 'd', 'a', 't', 'a', '.'}
	testInput3 = []rune{'.', 'l', 'e', 'x', 'i', 'n', 'g', ' ', 'd', 'a', 't', 'a', '.'}
)

// initState describes the StateFn to kick off the lexer / parser. It is also the default fallback StateFn
// for any other StateFn
func initState[C uint, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
recLoop:
	for {
		switch l.Next() {
		case '.':
			if l.Width()-1 > 0 {
				l.Prev()
				l.Emit((C)(tokenIdent))
				l.Next()
			}
			return statePeriod[C, T, I]
		case 0:
			break recLoop
		default:
		}
	}
	if l.Width() > 0 {
		l.Emit((C)(tokenIdent))
	}
	l.Emit((C)(tokenEOF))
	return nil
}

// statePeriod describes the StateFn to read a period in the content, emitting it as a period item
func statePeriod[C uint, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.Emit((C)(tokenPeriod))
	// look into the current rune
	switch l.Next() {
	case 0:
		l.Emit((C)(tokenEOF))
		return nil
	default:
		return initState[C, T, I]
	}
}

// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C uint, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.Emit((C)(tokenError))
	l.Ignore()
	return initState[C, T, I]
}

func TestNewLexer(t *testing.T) {
	l := lex.New(initState[uint, rune, lex.Item[uint, rune]], testInput1)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
}

func TestLexer(t *testing.T) {
	wants := "lexing.data."
	l := lex.New(initState[uint, rune, lex.Item[uint, rune]], testInput2)
	if l == nil {
		t.Errorf("output lexer should not be nil")
	}
	var items = []lex.Item[uint, rune]{}
	for {
		i := l.NextItem()
		items = append(items, i)
		if i.Type == tokenEOF {
			break
		}
	}

	if len(items) != 5 {
		t.Errorf("token slice length mismatch error: wanted %d ; got %d", 4, len(items))
	}
	if items[0].Type != tokenIdent {
		t.Errorf("unexpected token: wanted %v ; got %v", tokenIdent, items[0].Type)
	}
	if items[1].Type != tokenPeriod {
		t.Errorf("unexpected token: wanted %v ; got %v", tokenPeriod, items[1].Type)
	}
	if items[2].Type != tokenIdent {
		t.Errorf("unexpected token: wanted %v ; got %v", tokenIdent, items[2].Type)
	}
	if items[3].Type != tokenPeriod {
		t.Errorf("unexpected token: wanted %v ; got %v", tokenPeriod, items[3].Type)
	}
	if items[4].Type != tokenEOF {
		t.Errorf("unexpected token: wanted %v ; got %v", tokenPeriod, items[4].Type)
	}

	var output = new(strings.Builder)
	for _, i := range items {
		output.WriteString(string(i.Value))
	}

	if output.String() != wants {
		t.Errorf("unexpected output error: wanted %s ; got %s", wants, output.String())
	}
}

func TestNextItem(t *testing.T) {
	t.Run("LastChar", func(t *testing.T) {
		wants := "lexing data"
		l := lex.New(initState[uint, rune, lex.Item[uint, rune]], testInput1)
		if l == nil {
			t.Errorf("output lexer should not be nil")
		}
		i := l.NextItem()
		if i.Type != tokenIdent {
			t.Errorf("unexpected token type: wanted `ident` ; got %d -- item: %v", i.Type, string(i.Value))
		}
		if string(i.Value) != wants {
			t.Errorf("unexpected output value: wanted `%s` ; got `%s`", wants, string(i.Value))
		}
	})
	t.Run("MiddleChar", func(t *testing.T) {
		wants := "lexing"
		l := lex.New(initState[uint, rune, lex.Item[uint, rune]], testInput2)
		if l == nil {
			t.Errorf("output lexer should not be nil")
		}
		i := l.NextItem()
		if i.Type != tokenIdent {
			t.Errorf("unexpected token type: wanted `ident` ; got %d -- item: %v", i.Type, string(i.Value))
		}
		if string(i.Value) != wants {
			t.Errorf("unexpected output value: wanted `%s` ; got `%s`", wants, string(i.Value))
		}
	})
	t.Run("FirstChar", func(t *testing.T) {
		wants := "."
		l := lex.New(initState[uint, rune, lex.Item[uint, rune]], testInput3)
		if l == nil {
			t.Errorf("output lexer should not be nil")
		}
		i := l.NextItem()
		if i.Type != tokenPeriod {
			t.Errorf("unexpected token type: wanted `ident` ; got %d -- item: %v", i.Type, string(i.Value))
		}
		if string(i.Value) != wants {
			t.Errorf("unexpected output value: wanted `%s` ; got `%s`", wants, string(i.Value))
		}
	})
}

func TestPositionMethods(t *testing.T) {
	// input: `lexing data.`
	l := lex.New(initState[uint, rune, lex.Item[uint, rune]], testInput1)

	if l.Len() != 12 {
		t.Errorf("unexpected runes length: wanted %d ; got %d", 12, l.Len())
	}

	r := l.Next()
	if r != 'l' || r != l.PeekOffset(-1) {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "l", string(r))
	}
	if l.Start() != 0 {
		t.Errorf("unexpected start value: wanted %d ; got %d", 0, l.Start())
	}
	if l.Width() != 1 {
		t.Errorf("unexpected width value: wanted %d ; got %d", 1, l.Width())
	}
	if l.Pos() != 1 {
		t.Errorf("unexpected pos value: wanted %d ; got %d", 1, l.Pos())
	}
	r = l.Cur()
	if r != 'e' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "e", string(r))
	}

	l.Ignore()
	if l.Start() != 1 {
		t.Errorf("unexpected start value: wanted %d ; got %d", 1, l.Start())
	}

	l.Next()
	l.Next()
	l.Backup()
	if l.Pos() != 1 {
		t.Errorf("unexpected pos value: wanted %d ; got %d", 1, l.Pos())
	}
	if l.Width() != 0 {
		t.Errorf("unexpected width value: wanted %d ; got %d", 0, l.Width())
	}
	l.Head()

	r = l.Peek()
	if r != 'e' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "e", string(r))
	}

	r = l.Head()
	if r != 'l' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "l", string(r))
	}
	zero := l.Prev()
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	if l.Pos() != 0 || l.Width() != 0 {
		t.Errorf("unexpected pos/width value: wanted %d ; got %d / %d", 0, l.Pos(), l.Width())
	}
	l.Next()
	r = l.Prev()
	if r != 'l' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "l", string(r))
	}
	if l.Pos() != 0 || l.Width() != 0 {
		t.Errorf("unexpected pos/width value: wanted %d ; got %d / %d", 0, l.Pos(), l.Width())
	}

	r = l.Tail()
	if r != '.' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", ".", string(r))
	}
	l.Next() // accept last character
	zero = l.Next()
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	zero = l.Cur()
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	if l.Pos() != 12 || l.Width() != 1 {
		t.Errorf("unexpected pos/width value: wanted %d ; got %d / %d", 1, l.Pos(), l.Width())
	}
	zero = l.Peek()
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}

	r = l.Idx(10)
	if r != 'a' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "a", string(r))
	}
	if l.Start() != 10 || l.Width() != 0 {
		t.Errorf("unexpected start/width value: wanted %d ; got %d / %d", 0, l.Start(), l.Width())
	}

	zero = l.Idx(30)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	zero = l.Idx(-1)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}

	l.Head()
	r = l.Offset(2)
	if r != 'x' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "x", string(r))
	}
	if l.Start() != 0 || l.Pos() != 2 || l.Width() != 2 {
		t.Errorf("unexpected start/pos/width value: wanted %d / %d / %d ; got %d / %d / %d", 0, 2, 2, l.Start(), l.Pos(), l.Width())
	}

	r = l.Offset(-1)
	if r != 'e' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "e", string(r))
	}
	if l.Start() != 0 || l.Pos() != 1 || l.Width() != 1 {
		t.Errorf("unexpected start/pos/width value: wanted %d / %d / %d ; got %d / %d / %d", 0, 1, 1, l.Start(), l.Pos(), l.Width())
	}

	l.Offset(3) // idx 4
	l.Ignore()
	if l.Start() != 4 || l.Pos() != 4 || l.Width() != 0 {
		t.Errorf("unexpected start/pos/width value: wanted %d / %d / %d ; got %d / %d / %d", 4, 4, 0, l.Start(), l.Pos(), l.Width())
	}

	l.Offset(-2)
	if l.Start() != 2 || l.Pos() != 2 || l.Width() != 0 {
		t.Errorf("unexpected start/pos/width value: wanted %d / %d / %d ; got %d / %d / %d", 2, 2, 0, l.Start(), l.Pos(), l.Width())
	}
	zero = l.Offset(-5)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	if l.Start() != 2 || l.Pos() != 2 || l.Width() != 0 {
		t.Errorf("unexpected start/pos/width value: wanted %d / %d / %d ; got %d / %d / %d", 2, 2, 0, l.Start(), l.Pos(), l.Width())
	}
	zero = l.Offset(50)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	if l.Start() != 2 || l.Pos() != 2 || l.Width() != 0 {
		t.Errorf("unexpected start/pos/width value: wanted %d / %d / %d ; got %d / %d / %d", 2, 2, 0, l.Start(), l.Pos(), l.Width())
	}

	r = l.PeekIdx(2)
	if r != 'x' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "x", string(r))
	}
	zero = l.PeekIdx(-5)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	zero = l.PeekIdx(50)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	r = l.PeekOffset(2)
	if r != 'n' {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "n", string(r))
	}
	zero = l.PeekOffset(-5)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}
	zero = l.PeekOffset(50)
	if zero != 0 {
		t.Errorf("unexpected rune value: wanted `%s` ; got `%s`", "", string(zero))
	}

	extract := l.Extract(0, 3)
	if string(extract) != "lex" {
		t.Errorf("unexpected extracted string value: wanted `%s` ; got `%s`", "lex", string(extract))
	}

	extract = l.Extract(-5, 3)
	if string(extract) != "lex" {
		t.Errorf("unexpected extracted string value: wanted `%s` ; got `%s`", "lex", string(extract))
	}
	extract = l.Extract(7, 50)
	if string(extract) != "data." {
		t.Errorf("unexpected extracted string value: wanted `%s` ; got `%s`", "data.", string(extract))
	}
}

// acceptanceState describes the StateFn to kick off the lexer / parser with an acceptance func.
// It is also the default fallback StateFn for any other StateFn
func acceptanceState[C uint, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.AcceptRun(func(item T) bool {
		return item != '.' && item != 0
	})
	if l.Width() > 0 {
		l.Emit((C)(tokenIdent))
	}

	switch l.Next() {
	case '.':
		l.Prev()
		if l.Width() > 0 {
			l.Emit((C)(tokenIdent))
		}
		return acceptPeriod[C, T, I]
	case 0:
		l.Emit((C)(tokenEOF))
		return nil
	}
	return nil
}

func acceptPeriod[C uint, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	if l.Accept(func(item T) bool {
		return item == '.'
	}) {
		l.Emit((C)(tokenPeriod))
	}
	// verify we're standing on a period for testing purposes
	if l.Check(func(item T) bool {
		return item == '.'
	}) {
	}
	// verify we're not standing on EOF for testing purposes
	if l.Check(func(item T) bool {
		return item == 0
	}) {
		l.Emit((C)(tokenEOF))
		return nil
	}

	// accept a second period in the same go for testing purposes
	if l.Accept(func(item T) bool {
		return item == '.'
	}) {
		l.Emit((C)(tokenPeriod))
	}

	// look into the next rune
	switch l.Peek() {
	case 0:
		l.Emit((C)(tokenEOF))
		return nil
	default:
		return initState[C, T, I]
	}
}

func TestAcceptanceMethods(t *testing.T) {
	// input: `lexing.data.`
	wantsString := "lexing"
	wantsPeriod := "."
	wantsString2 := "data"
	l := lex.New(acceptanceState[uint, rune, lex.Item[uint, rune]], testInput2)

	i := l.NextItem()
	if i.Type != tokenIdent {
		t.Errorf("unexpected token type: wanted %s ; got %v", "tokenIdent", i.Type)
	}
	if string(i.Value) != wantsString {
		t.Errorf("unexpected token type: wanted %s ; got %s", wantsString, string(i.Value))
	}

	i = l.NextItem()
	if i.Type != tokenPeriod {
		t.Errorf("unexpected token type: wanted %s ; got %v", "tokenIdent", i.Type)
	}
	if string(i.Value) != wantsPeriod {
		t.Errorf("unexpected token type: wanted %s ; got %s", wantsPeriod, string(i.Value))
	}

	i = l.NextItem()
	if i.Type != tokenIdent {
		t.Errorf("unexpected token type: wanted %s ; got %v", "tokenIdent", i.Type)
	}
	if string(i.Value) != wantsString2 {
		t.Errorf("unexpected token type: wanted %s ; got %s", wantsString2, string(i.Value))
	}
	i = l.NextItem()
	if i.Type != tokenPeriod {
		t.Errorf("unexpected token type: wanted %s ; got %v", "tokenIdent", i.Type)
	}
	if string(i.Value) != wantsPeriod {
		t.Errorf("unexpected token type: wanted %s ; got %s", wantsPeriod, string(i.Value))
	}
}
