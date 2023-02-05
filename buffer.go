package lex

import (
	"io"

	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/gio"
)

const bufferInitSize = 1024

// LexBuffer implements the Lexer interface, by accepting a gio.Reader of any type
type LexBuffer[C comparable, T any] struct {
	input gio.Reader[T]
	buf   []T
	start int
	pos   int
	state StateFn[C, T]
	items chan Item[C, T]
}

var _ Lexer[uint8, any] = &LexBuffer[uint8, any]{}

func NewBuffer[C comparable, T any](
	initFn StateFn[C, T],
	input gio.Reader[T],
) *LexBuffer[C, T] {
	if input == nil {
		return nil
	}

	return &LexBuffer[C, T]{
		input: input,
		buf:   make([]T, 0, bufferInitSize),
		state: initFn,
		items: make(chan Item[C, T], 2),
	}
}

// NextItem processes the tokens sequentially, through the corresponding StateFn
//
// As each item is processed, it is returned to the Lexer by `Emit()`, and
// finally returned to the caller.
//
// Note that multiple calls to `NextItem()` should be made when tokenizing input data;
// usually in a for-loop while the output item is not EOF.
func (l *LexBuffer[C, T]) NextItem() Item[C, T] {
	var next Item[C, T]
	for {
		select {
		case next = <-l.items:
			return next
		default:
			if l.state != nil {
				l.state = l.state(l)
				continue
			}
			close(l.items)
			return next
		}
	}
}

// Emit pushes the set of units identified by token `itemType` to the items channel,
// that returns it in the NextItem() method.
//
// The emitted item will be a subsection of the input data slice, from the lexer's
// starting index to the current position index.
//
// It also sets the lexer's starting index to the current position index.
func (l *LexBuffer[C, T]) Emit(itemType C) {
	l.items <- Item[C, T]{
		Pos:   l.start,
		Type:  itemType,
		Value: l.buf[l.start:l.pos],
	}
	l.start = l.pos
}

// Ignore will set the starting point as the current position, ignoring any preceeding units
func (l *LexBuffer[C, T]) Ignore() {
	l.start = l.pos
}

// Backup will rewind the index for the width of the current item
func (l *LexBuffer[C, T]) Backup() {
	l.pos = l.start
}

// Width returns the size of the set of units ready to be emitted with a token
func (l *LexBuffer[C, T]) Width() int {
	return l.pos - l.start
}

// Start returns the current starting-point index for when an item is emitted
func (l *LexBuffer[C, T]) Start() int {
	return l.start
}

// Check passes the current token through the input `verifFn` function as a validator, returning
// its result
func (l *LexBuffer[C, T]) Check(verifFn func(item T) bool) bool {
	return verifFn(l.Cur())
}

// Accept passes the current token through the input `verifFn` function as a validator, returning
// its result
//
// If the validation passes, the cursor has moved one step forward (the unit was consumed)
//
// If the validation fails, the cursor rolls back one step
func (l *LexBuffer[C, T]) Accept(verifFn func(item T) bool) bool {
	if ok := verifFn(l.Next()); ok {
		return true
	}
	l.Prev()
	return false
}

// AcceptRun iterates through all following tokens, passing them through the input `verifFn`
// function as a validator
//
// Once it fails the verification, the cursor is rolledback once, leaving the caller at the unit
// that failed the verifFn
func (l *LexBuffer[C, T]) AcceptRun(verifFn func(item T) bool) {
	for verifFn(l.Next()) {
	}
	l.Prev()
}

func (l *LexBuffer[C, T]) consume(pos int) error {
	if n := pos - len(l.buf) + 1; n > 0 {
		for i := 0; i < n; i++ {
			var next = make([]T, 1)
			n, err := l.input.Read(next)
			if err != nil {
				return err
			}
			if n == 0 {
				return io.EOF
			}
			l.buf = append(l.buf, next[0])
		}
	}
	return nil
}

// Cur returns the same indexed item in the slice
//
// If the position is already over the size of the input, the zero-value EOF token is returned
func (l *LexBuffer[C, T]) Cur() T {
	if l.pos >= len(l.buf) {
		var eof T
		return eof
	}
	return l.buf[l.pos]
}

// Pos returns the current position in the cursor
func (l *LexBuffer[C, T]) Pos() int {
	return l.pos
}

// Len returns the current size of the underlying slice
func (l *LexBuffer[C, T]) Len() int {
	return len(l.buf)
}

// Next returns the current item in the slice as per the lexer's position, and
// increments the position by one.
//
// If the position is bigger or equal to the size of the input data, the position
// value is NOT incremented and the zero-value EOF token is returned
func (l *LexBuffer[C, T]) Next() T {
	l.pos++
	err := l.consume(l.pos)
	if err != nil {
		var eof T
		return eof
	}
	return l.buf[l.pos-1]
}

// Prev returns the previous item in the slice, while also decrementing the
// lexer's position.
//
// If the new position is less than zero, the position value is NOT decremented
// and the zero-value EOF token is returned
func (l *LexBuffer[C, T]) Prev() T {
	if l.pos-1 < 0 {
		var eof T
		return eof
	}
	l.pos--
	for l.pos < l.start {
		l.start--
	}
	err := l.consume(l.pos)
	if err != nil {
		var eof T
		return eof
	}
	return l.buf[l.pos]
}

// Peek returns the next indexed item without advancing the cursor
//
// If the next token overflows the input's index, the zero-value EOF token is returned
func (l *LexBuffer[C, T]) Peek() T {
	err := l.consume(l.pos)
	if err != nil {
		var eof T
		return eof
	}
	next := l.buf[l.pos]
	l.pos--
	return next
}

// Head returns to the beginning of the slice, setting both lexer's start and position
// values to zero
func (l *LexBuffer[C, T]) Head() T {
	l.pos = 0
	l.start = 0
	return l.buf[l.pos]
}

// Tail jumps to the end of the slice, setting both lexer's start and position values to
// the last item in the input
func (l *LexBuffer[C, T]) Tail() T {
	buf := gbuf.NewBuffer(l.buf)
	n, err := buf.ReadFrom(l.input)
	if err != nil || n == 0 {
		var eof T
		return eof
	}

	l.pos = len(l.buf) - 1
	l.start = len(l.buf) - 1
	return l.buf[l.pos]
}

// Idx jumps to the specific index `idx` in the slice
//
// If the input index is below 0, the zero-value EOF token is returned
// If the input index is greater than the size of the input, the
// zero-value EOF token is returned
func (l *LexBuffer[C, T]) Idx(idx int) T {
	if idx < 0 {
		var eof T
		return eof
	}
	if idx >= len(l.buf) {
		var eof T
		return eof
	}
	l.pos = idx
	if idx < l.start {
		l.start = idx
	}
	err := l.consume(l.pos)
	if err != nil {
		var eof T
		return eof
	}

	return l.buf[l.pos]
}

// Offset advances or rewinds `amount` steps in the slice, be it a positive or negative
// input.
//
// If the result offset is below 0, the zero-value EOF token is returned
// If the result offset is greater than the size of the slice, the
// zero-value EOF token is returned
func (l *LexBuffer[C, T]) Offset(amount int) T {
	if l.pos+amount < 0 {
		var eof T
		return eof
	}
	if l.pos+amount >= len(l.buf) {
		var eof T
		return eof
	}
	l.pos += amount
	if l.pos-l.start < 0 {
		l.start += amount
	}
	err := l.consume(l.pos)
	if err != nil {
		var eof T
		return eof
	}

	return l.buf[l.pos]
}

// PeekIdx returns the next indexed item without advancing the cursor,
// with the index `idx`
//
// If the input index is below 0, the zero-value EOF token is returned
// If the input index is greater than the size of the input, the
// zero-value EOF token is returned
func (l *LexBuffer[C, T]) PeekIdx(idx int) T {
	if idx >= len(l.buf) {
		var eof T
		return eof
	}
	if idx < 0 {
		var eof T
		return eof
	}
	err := l.consume(idx)
	if err != nil {
		var eof T
		return eof
	}

	return l.buf[idx]
}

// PeekOffset returns the next indexed item without advancing the cursor,
// with offset `amount`
//
// If the result offset is below 0, the zero-value EOF token is returned
// If the result offset is greater than the size of the slice, the
// zero-value EOF token is returned
func (l *LexBuffer[C, T]) PeekOffset(amount int) T {
	if l.pos+amount >= len(l.buf) || l.pos+amount < 0 {
		var eof T
		return eof
	}
	err := l.consume((l.pos + amount))
	if err != nil {
		var eof T
		return eof
	}

	return l.buf[l.pos+amount]
}

// Extract returns a slice from index `start` to index `end`
//
// If the input start index is below 0, the starting point will be set to zero
// If the input end index is greater than the size of the input, the
// ending point will be set to the size of the input.
func (l *LexBuffer[C, T]) Extract(start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end > len(l.buf) {
		end = len(l.buf)
	}
	return l.buf[start:end]
}
