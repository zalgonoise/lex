package protofile

import (
	_ "embed"
	"testing"

	"github.com/zalgonoise/gbuf"
	"github.com/zalgonoise/gio"
)

//go:embed testdata/all.proto
var protofile []byte

func TestParser(t *testing.T) {
	r := (gio.Reader[byte])(gbuf.NewReader(protofile))

	str, err := Run(r)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(str)
}
