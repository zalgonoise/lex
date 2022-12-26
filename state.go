package lex

// StateFn is a recursive function that updates the Lexer state according to a specific
// input token along the lexing / parsing action.
//
// It returns another StateFn or nil, as it consumes each token with a certain logic applied,
// passing along the lexing / parsing to the next StateFn
type StateFn[C comparable, T any, I Item[C, T]] func(l Lexer[C, T, I]) StateFn[C, T, I]
