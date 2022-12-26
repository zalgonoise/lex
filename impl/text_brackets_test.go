package impl

import "testing"

func TestRun(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		wants := `string with >>template<< in it`
		input := `string with {template} in it`
		out, err := Run(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if wants != out {
			t.Errorf("unexpected output error: wanted %s ; got %s", wants, out)
		}
	})
	t.Run("errored", func(t *testing.T) {
		wants := "string with :ERR:"
		input := `string with {template in it

	`

		out, err := Run(input)
		if err == nil {
			t.Errorf("expected error not to be nil")
		}
		if wants != out {
			t.Errorf("unexpected output error: wanted %s ; got %s", wants, out)
		}
	})
}
