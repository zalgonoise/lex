# lex

*a generic lexer written in Go*

_______________

## Overview

`lex` is a lexer for Go, based on the concept of the [`text/template`](https://pkg.go.dev/text/template) lexer, as a generic implementation. The logic behind this lexer is mostly based off of [Rob Pike](https://github.com/robpike)'s talk about [Lexical Scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE), which is also seen in the standard library (in [`text/template/parse/lex.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/text/template/parse/lex.go)).


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
type StateFn[C comparable, T any, I Item[C, T]] func(l Lexer[C, T, I]) StateFn[C, T, I]
```

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
type Lexer[C comparable, T any, I Item[C, T]] interface {

	// Cursor navigates through a slice in a controlled manner, allowing the
	// caller to move forward, backwards, and jump around the slice as they need
	cur.Cursor[T]

	// NextItem processes the tokens sequentially, through the corresponding StateFn
	//
	// As each item is processed, it is returned to the Lexer by `Emit()`, and
	// finally returned to the caller.
	//
	// Note that multiple calls to `NextItem()` should be made when tokenizing input data;
	// usually in a for-loop while the output item is not EOF.
	NextItem() I

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
```

## Implementing

**Note**: *Example and tests can be found in the [`impl`](./impl/) directory.*

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


> `initState` absorbs all characters until it hits a `{` or EOF. Then, if the following 
> character is a `{`, it returns the `stateLBRACE` routine. If it hits EOF, it will 
> return a EOF token and a nil `StateFn`.
>
> The checks for `l.Width() > 0` ensures that an existing *stack* is being pushed before 
> advancing to the next token in a different procedure (e.g., consider all identifier tokens
> before going into the `stateBRACE` routine)


```go
// initState describes the StateFn to kick off the lexer. It is also the default fallback StateFn
// for any other StateFn
func initState[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.AcceptRun(func(item T) bool {
		return item != '{' && item != 0
	})

	switch l.Next() {
	case '{':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenIDENT))
		}
		l.Ignore()
		return stateLBRACE[C, T, I]
	default:
		if l.Width() > 0 {
			l.Emit((C)(TokenIDENT))
		}
		l.Emit((C)(TokenEOF))
		return nil
	}
}
```

> `stateLBRACE` takes care of the *template text* tokenization. It will absorb the text
> before and after a `{` and `}` respectively, emitting a *template* token when done.
>
> If the template is not closed, it will return a `stateError` routine, as a template should
> always be closed. This examplifies error handling in a simple manner.
>
> Note the calls to `l.Check()`, verifying if the current token is `{`. Since this 
> implementation does not tokenize these symbols (only the template text), it is skipped 
> with `l.Next()` and `l.Ignore()`. Otherwise, one should emit these items as a specific 
> token

```go
// stateLBRACE describes the StateFn to read the template content, emitting it as a template item
func stateLBRACE[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	if l.Check(func(item T) bool {
		return item == '{'
	}) {
		l.Next() // skip this symbol
		l.Ignore()
	}

	l.AcceptRun(func(item T) bool {
		return item != '}' && item != 0
	})

	switch l.Next() {
	case '}':
		if l.Width() > 0 {
			l.Prev()
			l.Emit((C)(TokenTEMPL))
		}
		l.Next() // skip this symbol
		l.Ignore()
		return initState[C, T, I]
	default:
		return stateError[C, T, I]
	}
}
```

> `stateError` basically signals the Lexer that there was an error (with no specific 
> metadata) when analyzing this section. It will return the `initState StateFn`, but it could simply return nil to terminate the analysis.

```go
// stateError describes an errored state in the lexer / parser, ignoring this set of tokens and emitting an
// error item
func stateError[C TextToken, T rune, I lex.Item[C, T]](l lex.Lexer[C, T, I]) lex.StateFn[C, T, I] {
	l.Emit((C)(TokenError))
	return initState[C, T, I]
}
```


Finally, the Lexer can be initialized with the `New` function, by supplying the initial `StateFn` and the input data.

> TextTemplateLexer is just imposing the usage of `TextToken` as a comparable token, and 
> `rune` as a lexeme type when creating a Lexer, along with setting its initial `StateFn` to 
> `initState` 
>
> Its signature is super big, but when called, it's as simple as `l := TextTemplateLexer(input)`

```go
// TextTemplateLexer creates a text template lexer based on the input slice of runes
func TextTemplateLexer[C TextToken, T rune, I lex.Item[C, T]](input []T) lex.Lexer[C, T, I] {
	return lex.New(initState[C, T, I], input)
}
```

Lastly, as helpers, the developer can also implement a converter for the generic `lex.Item[C, T, I]` returned from `Lexer.NextItem()`, into the `Item` type of their choice. Also, a `Run[T any](input T)` or `Execute[T any](input T)` could be added, that loops through  the `Lexer.NextItem()` calls until it reaches an EOF token. To convert the output Items into a string, I chose to implement `fmt.Stringer` -- but this should be the Parser's responsibility.

> `toTemplateItem` is a simple converter to avoid type-casting the generic Item type to our own

```go
// toTemplateItem converts a lex.Item type to TemplateItem
func toTemplateItem[C TextToken, T rune](i lex.Item[C, T]) TemplateItem[C, T] {
	return (TemplateItem[C, T])(i)
}
```

> `Run` takes care of the `l.NextItem()` loop, as it is aggregating Items. It will also take care of *parsing* the tokens too, as the sets of runes will be presented as a string (or error if found).

```go
// Run takes in a string `s`, processes it for templates, and returns the processed string and an error
func Run(s string) (string, error) {
	l := TextTemplateLexer([]rune(s))
	var sb = new(strings.Builder)
	for {
		i := l.NextItem()
		sb.WriteString(toTemplateItem(i).String())

		switch i.Type {
		case 0:
			return sb.String(), nil
		case TokenError:
			return sb.String(), fmt.Errorf("failed to parse token")
		}
	}
}
```

> For context, `TemplateItem.String()` will simply switch on the token type in the item; 
> performing any manipulation or processing necessary and finally returning a string.
> 
> Note how the (nested) generic types force casting the token type as a generic type `C`:

```go
// String implements fmt.Stringer; which is processing each TemplateItem as a string
func (t TemplateItem[C, I]) String() string {
	switch t.Type {
	case C(TokenIDENT):
		var rs = make([]rune, len(t.Value), len(t.Value))
		for idx, r := range t.Value {
			rs[idx] = (rune)(r)
		}
		return string(rs)
	case C(TokenTEMPL):
		var rs = make([]rune, len(t.Value), len(t.Value))
		for idx, r := range t.Value {
			rs[idx] = (rune)(r)
		}
		return ">>" + string(rs) + "<<"
	case C(TokenError):
		return ":ERR:"
	case C(TokenEOF):
		return "" // placeholder action for EOF tokens
	}
	return ""
}
```


## Usage

From this point forward, once the Lexer's logic is defined (and well-tested), the users can simply consume the Items returned by repeated calls of the `Lexer.NextItem()` method, until EOF is reached. This is summarized in the last section's `Run()` function example, in this snippet:

```go
// Run takes in a string `s`, processes it for templates, and returns the processed string and an error
func Run(s string) (string, error) {
	l := TextTemplateLexer([]rune(s))
	for {
		i := l.NextItem()
		switch i.Type {
		    // (...)
        }
	}
}
```

In this particular example, the Run function takes care of the analysis and also the processing (lexing **and** parsing), because it's a super simple example and it's easier to visualize with an A-B comparison. However, the `Run()` function could simply return the slice of Item collected in the loop.

## Benchmarks

Performance is critical in a lexer or parser. I will add benchmarks (and performance improvements) very soon :)