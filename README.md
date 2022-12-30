# lex

*a generic lexer library written in Go*

_______________

## Concept

`lex` is a lexer library for Go, based on the concept of the [`text/template`](https://pkg.go.dev/text/template) lexer, as a generic implementation. The logic behind this lexer is mostly based off of [Rob Pike](https://github.com/robpike)'s talk about [Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE), which is also seen in the standard library (in [`text/template/parse/lex.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/text/template/parse/lex.go)).

The lexer is a state-machine that analyzes input data (split into single units) by traversing the slice and classifying *blobs* of the slice (as lexemes) with a certain token. The lexer emits items, which are composed of three main elements:
- the starting position in the slice where the lexeme starts
- the (comparable type) token classifying this piece of data
- a slice containing the actual data 

The lexer is usually in-sync with a parser (also a state-machine, running in tandem with a lexer), that will consume the items emited by the lexer to build a parse tree. The parse tree, as the name implies, is a tree graph data structure that will layout the received tokens with a certain path / structure, with some logic in mind. It is finally able to output the processed tree as an output type, configurable by the developer, too.

## Why generics?

Generics are great for when there is a solid algorithm that serves for many types, and can be abstracted enough to work without major workarounds; and this approach to a lexer / parser is very straight-forward and yet so simple (the Go way). Of course when I refer [Rob Pike](https://github.com/robpike)'s talk about lexers I am aware that the context is parsing text (for templating). The *approach with generics* will limit the potential that shines in the original implementation, one way or the other (simply with EOF being a zero value, for example -- zero types should not be used for this).

But all in all, it was a great exercise to practice using generics. Maybe I will just use this library once or twice, maybe it will be actually useful for some. I am just in it for the ride. :)

## Overview

The idea behind implementing a generic algorithm for a lexer came from trying to build a graph (data structure) representing the logic blocks in a Go file. Watching the talk above was a breath of fresh air when it came to the design of the lexer and its simple approach. So, it would be nice to leverage this algorithm for the Go code graph idea from before. By making the logic generic, one could implement an `Item` type to hold a defined token type, and a set of (any type of) values and `StateFn` state-functions to tokenize input data. In concept this works for any given type, as the point is to label elements of a slice with identifying tokens, that will be processed into a parse tree (with a specific parser implementation).

Caveats are precisely using very *open* types for this implementation. The `text/template` lexer will, for example, define its EOF token as `-1` -- a constant found in the [`lex.go` file](https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/text/template/parse/lex.go;l=93). For this implementation, the lexer will return a zero-value token, so the caller should prepare their token types considering that the zero value will be reserved for EOF. Scrolling through the input will use the `pos int` position, and will not have a width -- because the lexer will consume the input as a list of the defined data type. Concerns about the width of a unit need to be handled in the types or `StateFn` implementation, not the Lexer.

Additionally, as it is exposed as an interface (which is not set in stone, though), it introduces a few helper methods to either validate input data, navigate through the index, and controlling the cursor of the slice. It is implementing the [`cur.Cursor[T any] interface`](https://github.com/zalgonoise/cur).

## Installation 

> Note: this library is not ready out-of-the box! You will need to implement your own `StateFn` state-functions with defined types. This repo will expose simple examples to understand the flow of the lexer, below.

You can add this library to your Go project with a `go get` command:

```
go get github.com/zalgonoise/lex
```

## Features

### Entities

#### Lexer

The Lexer is a state-machine that keeps track of the generated Items as the input data is consumed and labled with tokens. It's part cursor, part controller/verifier (within the `StateFn`s), but its main job is keep the state-functions running as the items are consumed, returning lexical items to the caller as they are generated.

The Lexer should be accompanied with a Parser, that consumes the tokenized Items to build a parse tree.

This Lexer exposes methods that should be perceived as utilities for the caller when building `StateFn`s. In reality, when actually *running* the Lexer, the caller will loop through its `NextItem()` method until it hits an EOF token.

The methods in the Lexer will be covered individually below, as well as the design decisions when writing it this way.

```go
// Lexer describes the behavior of a lexer / parser with cursor capabilities
//
// Once spawned, it will consume all tokens as its `NextItem()` method is called,
// returning processed `Item[C, T]` as it goes
//
// Its `Emit()` method pushes items into the stack to be returned, and its `Accept()`,
// `Check()` and `AcceptRun()` methods act as verifiers for a (set of) token(s)
type Lexer[C comparable, T any] interface {

	// Cursor navigates through a slice in a controlled manner, allowing the
	// caller to move forward, backwards, and jump around the slice as they need
	cur.Cursor[T]

	// Emitter describes the behavior of an object that can emit lex.Items
	//
	// It contains a single method, NextItem, that processes the tokens sequentially
	// through the corresponding StateFn
	//
	// As each item is processed, it is returned to the Lexer by `Emit()`, and
	// finally returned to the caller.
	//
	// Note that multiple calls to `NextItem()` should be made when tokenizing input data;
	// usually in a for-loop while the output item is not EOF.
	Emitter[C, T]

	// Emit pushes the set of units identified by token `itemType` to the items channel,
	// that returns it in the NextItem() method.
	//
	// The emitted item will be a subsection of the input data slice, from the lexer's
	// starting index to the current position index.
	//
	// It also sets the lexer's starting index to the current position index.
	Emit(itemType C)

	// Ignore will set the starting point as the current position, ignoring any preceeding units
	Ignore()

	// Backup will rewind the index for the width of the current item
	Backup()

	// Width returns the size of the set of units ready to be emitted with a token
	Width() int

	// Start returns the current starting-point index for when an item is emitted
	Start() int

	// Check passes the current token through the input `verifFn` function as a validator, returning
	// its result
	Check(verifFn func(item T) bool) bool

	// Accept passes the current token through the input `verifFn` function as a validator, returning
	// its result
	//
	// If the validation passes, the cursor has moved one step forward (the unit was consumed)
	//
	// If the validation fails, the cursor rolls back one step
	Accept(verifFn func(item T) bool) bool

	// AcceptRun iterates through all following tokens, passing them through the input `verifFn`
	// function as a validator
	//
	// Once it fails the verification, the cursor is rolledback once, leaving the caller at the unit
	// that failed the verifFn
	AcceptRun(verifFn func(item T) bool)
}

// Emitter describes the behavior of an object that can emit lex.Items
type Emitter[C comparable, T any] interface {
	// NextItem processes the tokens sequentially, through the corresponding StateFn
	//
	// As each item is processed, it is returned to the Lexer by `Emit()`, and
	// finally returned to the caller.
	//
	// Note that multiple calls to `NextItem()` should be made when tokenizing input data;
	// usually in a for-loop while the output item is not EOF.
	NextItem() Item[C, T]
}
```

#### Item

An Item is an object holding a token and a set of values (lexemes) corresponding to that token. It is a key-value data structure, where the value-half is a slice of any type -- which could be populated with any number of items.

Items are created as the lexer tokenizes symbols / units from the input data, and returned to the caller on each `Lexer.NextItem()` call. 

```go
// Item represents a set of any type of tokens identified by a comparable type
type Item[T comparable, V any] struct {
	Type  T
	Value []V
}
```

#### StateFn

A StateFn (state-function) describes a recursive function that will consume each unit of the input data, and either take action, emit tokens, or to finally pass along the data analysis action to a different StateFn. These functions will be implemented by the consumer of the library, describing the steps (their) lexer will take when consuming data. Examples further below.

The point to the StateFn is to keep the Lexer in control of the state, and the StateFn controlling the cursor and actions the Lexer needs to take. This will allow the biggest flexibility possible, as different Lexer StateFn will derive different flavors for the lexer. Thinking of it as a Markdown lexer, the StateFn would define which and how Markdown tokens are classified, emitted, or even ignored.


```go
// StateFn is a recursive function that updates the Lexer state according to a specific
// input token along the lexing / parsing action.
//
// It returns another StateFn or nil, as it consumes each token with a certain logic applied,
// passing along the lexing / parsing to the next StateFn
type StateFn[C comparable, T any] func(l Lexer[C, T]) StateFn[C, T]
```

## Implementing

**Note**: *Example and tests can be found in the [`impl`](./impl/) directory; from the lexer to the parser*

________

### Token type

Implementing a Lexer requires considering the format of the input data and how it can be tokenized. For this example, the input data is a string, where the lexeme units will be runes.

> The `TemplateItem` will be a comparable (unique) TextToken type, where the lexemes will be runes

```go
// TemplateItem represents the lex.Item for a runes lexer based on TextToken identifiers
type TemplateItem[C TextToken, I rune] lex.Item[C, I]
```

For this, the developer needs to define a token type (with an enumeration of expected tokens, where the zero-value for the type is EOF).

> A set of expected tokens are enumerated. In this case the text template will take text
> between single braces (like `{this}`), and ...replace the braces with double angle-brackets
> (like `>>this<<`). Not very fancy but serves as an example.

```go
// TextToken is a unique identifier for this text template implementation
type TextToken int

const (
	TokenEOF TextToken = iota
	TokenError
	TokenIDENT
	TokenTEMPL
	TokenLBRACE
	TokenRBRACE
)
```


After defining the type, a (set of) `StateFn`(s) need to be created, in context of the input data and how it should be tokenized. Each `StateFn` will hold the responsibility of tokenizing a certain lexeme, and each `StateFn` will have a different flow and responsibility.

### Lexer and state functions

> `initState` switches on the next lexable unit's value, to either emit an item or simply return a new state. This state should be able to listen to all types of (supported) symbols since this example supports so (a user could start a template in the very first char, and end it on the last one)
>
> The checks for `l.Width() > 0` ensures that an existing *stack* is being pushed before 
> advancing to the next token in a different procedure (e.g., consider all identifier tokens
> before going into the `stateBRACE` routine)


```go
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
```

> `stateIDENT` absorbs all text characters until it hits a `{`, `}` or EOF. Then, if the following 
> character is a `{`, or a `}` it returns the `stateLBRACE` or `stateRBRACE` routine, respectively. 
> If it hits EOF, it will return a EOF token and a nil `StateFn`.


```go
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
```

> `stateLBRACE` tokenizes the `{` character, returning the initial state after skipping this character

```go
// stateLBRACE describes the StateFn to check for and emit an LBRACE token
func stateLBRACE[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.Next() // skip this symbol
	l.Emit((C)(TokenLBRACE))
	return initState[C, T]
}
```

> Similarly, `stateRBRACE` tokenizes the `}` character:

```go
// stateRBRACE describes the StateFn to check for and emit an RBRACE token
func stateRBRACE[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.Next() // skip this symbol
	l.Emit((C)(TokenRBRACE))
	return initState[C, T]
}
```

> Finally `stateError` tokenizes an error if found (none in this lexer's example)

```go
// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C TextToken, T rune](l lex.Lexer[C, T]) lex.StateFn[C, T] {
	l.Backup()
	l.Prev() // mark the previous char as erroring token
	l.Emit((C)(TokenError))
	return initState[C, T]
}
```

### Parser

#### Parse functions

> Just like the lexer, start by defining a top-level ParseFn that will scan for all expected tokens
>
> This function will peek into the next item from the lexer and return the appropriate ParseFn before actually consuming the token

```go
// initParse describes the ParseFn to kick off the parser. It is also the default fallback 
// for any other ParseFn
func initParse[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	for t.Peek().Type != C(TokenEOF) {
		switch t.Peek().Type {
		case (C)(TokenIDENT):
			return parseText[C, T]
		case (C)(TokenLBRACE), (C)(TokenRBRACE):
			return parseTemplate[C, T]
		}
	}
	return nil
}
```

> `parseText` simply consumes the item as a new node under the current.

```go
// parseText consumes the next item as a text token, creating a node for it under the
// current one in the tree. 
func parseText[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	t.Node(t.Next())
	return initParse[C, T]
}
```

> `parseTemplate` is a state where we're about to consume either a `{` or a `}`. 
>
> For `{` tokens, a template Node is created, as it will be a wrapper for one or more text or template items. Returns the initial state.
>
> For `}` tokens, the node that is parent to the `{` is set as the current position (closing the template)

```go
// parseTemplate creates a node for a template item, for which it expects both a text item edge
// that which also needs to contain an end-template edge.
//
// If it encounters a `}` token to close the template, it sets the position up three levels
// (back to the template's parent)
func parseTemplate[C TextToken, T rune](t *parse.Tree[C, T]) parse.ParseFn[C, T] {
	switch t.Peek().Type {
	case (C)(TokenLBRACE):
		t.Set(t.Parent())
		t.Node(t.Next())
	case (C)(TokenRBRACE):
		t.Node(t.Next())
		t.Set(t.Parent().Parent.Parent)
	}
	return initParse[C, T]
}
```

#### Process functions

> `processFn` is the *top-level* processor function, that will consume the nodes in the Tree.
>
> It will use a strings.Builder to create the returned string, and iterate through the Tree's
> root Node's edges and switching on its Type.
>
> The content written to the strings.Builder comes from the appropriate `NodeFn` for the Node type.

```go
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
```

> for text it's straight-forward, it just casts the T-type values as rune, and returns a string value of it

```go
// processText converts the T-type items into runes, and returns a string value of it
func processText[C TextToken, T rune, R string](n *parse.Node[C, T]) (R, error) {
	var val = make([]rune, len(n.Value), len(n.Value))
	for idx, r := range n.Value {
		val[idx] = (rune)(r)
	}
	return (R)(val), nil
}
```

> for templates, a few checks need to be made -- in this particular example it is to ensure that templates are terminated.
> 
> the `processTemplate ProcessFn` does that exactly -- it replaces the wrapper text with the appropriate content, adds in the text in the next node, and looks into that text node's edges for a `}` item (to mark the template as closed). Otherwise returns an error:

```go
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
```

#### Wrapper

> Perfect! Now all components are wired-up among themselves, and it just needs a simple entrypoint function
>
> For this, we can use the template `Parse` function to run it all at once:

```go
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
```

## Benchmarks

Performance is critical in a lexer or parser. I will add benchmarks (and performance improvements) very soon :)