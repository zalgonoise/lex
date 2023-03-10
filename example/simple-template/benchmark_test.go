package impl

import (
	"testing"

	"github.com/zalgonoise/lex"
)

func BenchmarkLexer(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		input := []rune(`with {tmpl}.`)
		var lexeme lex.Item[TextToken, rune]
		var eof lex.Item[TextToken, rune]

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l := lex.New(initState[TextToken, rune], input)
			for {
				lex := l.NextItem()
				if lex.Type == eof.Type {
					break
				}
				lexeme = lex
			}
		}
		_ = lexeme
	})
	b.Run("Complex", func(b *testing.B) {
		input := []rune(`string with {template} in it even { in {twice} out } in a row, or {even} { more {examples} if necessary}.`)
		var lexeme lex.Item[TextToken, rune]
		var eof lex.Item[TextToken, rune]

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			l := lex.New(initState[TextToken, rune], input)
			for {
				lex := l.NextItem()
				if lex.Type == eof.Type {
					break
				}
				lexeme = lex
			}
		}
		_ = lexeme
	})
}
