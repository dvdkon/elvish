package layout

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
)

func TestWriteListing(t *testing.T) {
	b := ui.NewBufferBuilder(10)
	WriteListing(
		b, " LIST ", "f",
		"line 1",
		"line 2", Selected,
		"line 3")
	buf := b.Buffer()
	wantBuf := ui.NewBufferBuilder(10).
		WriteStyled(ModeLine(" LIST ", true)).
		WritePlain("f").SetDotHere().
		Newline().WritePlain("line 1").
		Newline().WriteString("line 2    ", "inverse").
		Newline().WritePlain("line 3").
		Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Buf differs")
		t.Logf("Got: %s", buf.TTYString())
		t.Logf("Want: %s", wantBuf.TTYString())
	}
}