package impl

import (
	"github.com/zalgonoise/gio"
	"github.com/zalgonoise/lex"
	"github.com/zalgonoise/parse"
)

// Run parses the input templated data (a string as []rune), returning
// a processed string and an error
func Run[C TextToken, T rune, R string](s gio.Reader[T]) (R, error) {
	var rootEOF C
	l := (lex.Emitter[C, T])(lex.NewBuffer(initState[C, T], s))
	t := parse.New(l, initParse[C, T], rootEOF)
	t.Parse()
	return processFn[C, T, R](t)
}
