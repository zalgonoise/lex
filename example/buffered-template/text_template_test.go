package impl

import (
	"testing"

	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/gio"
)

func TestRun(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		wants := `string with >>template<< in it even >> in >>twice<< out << in a row.`
		input := `string with {template} in it even { in {twice} out } in a row.`
		buf := (gio.Reader[rune])(gbuf.NewReader([]rune(input)))
		out, err := Run(buf)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if wants != out {
			t.Errorf("unexpected output error: wanted \n\n%s\n ; got \n\n%s\n", wants, out)
		}
	})

	t.Run("complex", func(t *testing.T) {
		wants := `string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row. string with >>template<< in it even >> in >>twice<< out << in a row.`
		input := `string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row. string with {template} in it even { in {twice} out } in a row.`
		buf := (gio.Reader[rune])(gbuf.NewReader([]rune(input)))
		out, err := Run(buf)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if wants != out {
			t.Errorf("unexpected output error: wanted \n\n%s\n ; got \n\n%s\n", wants, out)
		}
	})

	t.Run("errored", func(t *testing.T) {
		wants := "string with "
		wantsErr := "parse error on line: 0"
		input := `string with {template in it`

		buf := (gio.Reader[rune])(gbuf.NewReader([]rune(input)))
		out, err := Run(buf)
		if err == nil {
			t.Errorf("expected error not to be nil")
		}
		if wantsErr != err.Error() {
			t.Errorf("unexpected output error: wanted %s ; got %s", wantsErr, err.Error())
		}
		if wants != out {
			t.Errorf("unexpected output error: wanted \n\n%s\n ; got \n\n%s\n", wants, out)
		}
	})
}
