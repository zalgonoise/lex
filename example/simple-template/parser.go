package impl

import "github.com/zalgonoise/parse"

// Run parses the input templated data (a string as []rune), returning
// a processed string and an error
func Run[C TextToken, T rune, R string](s []T) (R, error) {
	return parse.Run(
		s,
		initState[C, T],
		initParse[C, T],
		processFn[C, T, R],
	)
}
