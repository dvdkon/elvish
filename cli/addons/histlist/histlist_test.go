package histlist

import (
	"errors"
	"strings"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

func setup() (*cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	app := cli.NewApp(tty)
	codeCh, _ := app.ReadCodeAsync()
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

func TestStart_NoStore(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{})
	wantNotesBuf := ui.NewBufferBuilder(80).WritePlain("no history store").Buffer()
	if !ttyCtrl.VerifyNotesBuffer(wantNotesBuf) {
		t.Errorf("Wanted notes buffer not shown")
		t.Logf("Want: %s", wantNotesBuf.TTYString())
		t.Logf("Last: %s", ttyCtrl.LastNotesBuffer().TTYString())
	}
}

type faultyStore struct{}

var mockError = errors.New("mock error")

func (s faultyStore) AllCmds() ([]histutil.Entry, error) { return nil, mockError }

func TestStart_StoreError(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{Store: faultyStore{}})
	wantNotesBuf := ui.NewBufferBuilder(80).WritePlain("db error: mock error").Buffer()
	if !ttyCtrl.VerifyNotesBuffer(wantNotesBuf) {
		t.Errorf("Wanted notes buffer not shown")
		t.Logf("Want: %s", wantNotesBuf.TTYString())
		t.Logf("Last: %s", ttyCtrl.LastNotesBuffer().TTYString())
	}
}

func TestStart_OK(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	_ = ttyCtrl
	defer cleanup()

	store := histutil.NewMemoryStore()
	store.AddCmd(histutil.Entry{Text: "foo", Seq: 0})
	store.AddCmd(histutil.Entry{Text: "bar", Seq: 1})
	store.AddCmd(histutil.Entry{Text: "baz", Seq: 2})
	Start(app, Config{Store: store})

	// Test UI.
	wantBuf := ui.NewBufferBuilder(80).
		// empty codearea
		Newline().
		// combobox codearea
		WriteStyled(styled.MakeText("HISTLIST",
			"bold", "lightgray", "bg-magenta")).
		WritePlain(" ").
		SetDotToCursor().
		// unselected entries
		Newline().WritePlain("   0 foo").
		Newline().WritePlain("   1 bar").
		// last entry is selected
		Newline().WriteStyled(
		styled.MakeText("   2 baz"+strings.Repeat(" ", 72), "inverse")).
		Buffer()
	if !ttyCtrl.VerifyBuffer(wantBuf) {
		t.Errorf("Wanted buffer not shown")
		t.Logf("Want: %s", wantBuf.TTYString())
		t.Logf("Last buffer: %s", ttyCtrl.LastBuffer().TTYString())
	}

	// Test filtering.
	ttyCtrl.Inject(term.K('b'))
	wantBuf = ui.NewBufferBuilder(80).
		// empty codearea
		Newline().
		// combobox codearea
		WriteStyled(styled.MakeText("HISTLIST",
			"bold", "lightgray", "bg-magenta")).
		WritePlain(" b").
		SetDotToCursor().
		// unselected entries
		Newline().WritePlain("   1 bar").
		// last entry is selected
		Newline().WriteStyled(
		styled.MakeText("   2 baz"+strings.Repeat(" ", 72), "inverse")).
		Buffer()
	if !ttyCtrl.VerifyBuffer(wantBuf) {
		t.Errorf("Wanted buffer not shown")
		t.Logf("Want: %s", wantBuf.TTYString())
		t.Logf("Last buffer: %s", ttyCtrl.LastBuffer().TTYString())
	}

	// Test accepting.
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBuf = ui.NewBufferBuilder(80).
		// codearea now contains selected entry
		WritePlain("baz").SetDotToCursor().Buffer()
	if !ttyCtrl.VerifyBuffer(wantBuf) {
		t.Errorf("Wanted buffer not shown")
		t.Logf("Want: %s", wantBuf.TTYString())
		t.Logf("Last buffer: %s", ttyCtrl.LastBuffer().TTYString())
	}

	// Test accepting when there is already some text.
	store.AddCmd(histutil.Entry{Text: "baz2"})
	Start(app, Config{Store: store})
	ttyCtrl.Inject(term.K(ui.Enter))
	wantBuf = ui.NewBufferBuilder(80).
		WritePlain("baz").
		// codearea now contains newly inserted entry on a separate line
		Newline().WritePlain("baz2").SetDotToCursor().Buffer()
	if !ttyCtrl.VerifyBuffer(wantBuf) {
		t.Errorf("Wanted buffer not shown")
		t.Logf("Want: %s", wantBuf.TTYString())
		t.Logf("Last buffer: %s", ttyCtrl.LastBuffer().TTYString())
	}
}