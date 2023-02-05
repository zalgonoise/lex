package impl

import (
	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/parse"
)

// Run parses the input templated data (a string as []rune), returning
// a processed string and an error
func Run[C TextToken, T rune, R string](s []T) (R, error) {
	var rootEOF C

	reader := (gio.Reader[T])(gbuf.NewReader(s))
	l := (lex.Lexer[C, T])(lex.NewBuffer(initState[C, T], reader))
	t := parse.New(l, initParse[C, T], rootEOF)
	t.Parse()
	return processFn[C, T, R](t)
}
