package cliedit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

func initAPI(app *cli.App, ev *eval.Evaler, ns eval.Ns) {
	initMaxHeight(app, ns)
	initBeforeReadline(app, ev, ns)
	initAfterReadline(app, ev, ns)
	initInsert(app, ev, ns)

	initMiscBuiltins(app, ns)
	initBufferBuiltins(app, ns)
}

func initMaxHeight(app *cli.App, ns eval.Ns) {
	maxHeight := newIntVar(-1)
	app.AppSpec.MaxHeight = func() int { return maxHeight.GetRaw().(int) }
	ns.Add("max-height", maxHeight)
}

func initBeforeReadline(app *cli.App, ev *eval.Evaler, ns eval.Ns) {
	hook := newListVar(vals.EmptyList)
	ns["before-readline"] = hook
	app.AppSpec.BeforeReadline = func() {
		i := -1
		hook := hook.GetRaw().(vals.List)
		for it := hook.Iterator(); it.HasElem(); it.Next() {
			i++
			name := fmt.Sprintf("$before-readline[%d]", i)
			fn, ok := it.Elem().(eval.Callable)
			if !ok {
				// TODO(xiaq): This is not testable as it depends on stderr.
				// Make it testable.
				diag.Complainf("%s not function", name)
				continue
			}
			// TODO(xiaq): This should use stdPorts, but stdPorts is currently
			// unexported from eval.
			ports := []*eval.Port{
				{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr}}
			fm := eval.NewTopFrame(ev, eval.NewInternalSource(name), ports)
			fm.Call(fn, eval.NoArgs, eval.NoOpts)
		}
	}
}

func initAfterReadline(app *cli.App, ev *eval.Evaler, ns eval.Ns) {
	hook := newListVar(vals.EmptyList)
	ns["after-readline"] = hook
	app.AppSpec.AfterReadline = func(code string) {
		i := -1
		hook := hook.GetRaw().(vals.List)
		for it := hook.Iterator(); it.HasElem(); it.Next() {
			i++
			name := fmt.Sprintf("$after-readline[%d]", i)
			fn, ok := it.Elem().(eval.Callable)
			if !ok {
				// TODO(xiaq): This is not testable as it depends on stderr.
				// Make it testable.
				diag.Complainf("%s not function", name)
				continue
			}
			// TODO(xiaq): This should use stdPorts, but stdPorts is currently
			// unexported from eval.
			ports := []*eval.Port{
				{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr}}
			fm := eval.NewTopFrame(ev, eval.NewInternalSource(name), ports)
			fm.Call(fn, []interface{}{code}, eval.NoOpts)
		}
	}
}

func initInsert(app *cli.App, ev *eval.Evaler, ns eval.Ns) {
	/*
		abbr := vals.EmptyMap
		abbrVar := vars.FromPtr(&abbr)
		app.CodeArea.Abbreviations = makeMapIterator(abbrVar)

		binding := newBindingVar(emptyBindingMap)
		app.CodeArea.OverlayHandler = newMapBinding(app, ev, binding)

		quotePaste := newBoolVar(false)
		app.CodeArea.QuotePaste = func() bool { return quotePaste.GetRaw().(bool) }

		ns.AddNs("insert", eval.Ns{
			"abbr":        abbrVar,
			"binding":     binding,
			"quote-paste": quotePaste,
		})
	*/
}

func makeMapIterator(mv vars.PtrVar) func(func(a, b string)) {
	return func(f func(a, b string)) {
		for it := mv.GetRaw().(hashmap.Map).Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			ks, kok := k.(string)
			vs, vok := v.(string)
			if !kok || !vok {
				continue
			}
			f(ks, vs)
		}
	}
}

func newIntVar(i int) vars.PtrVar            { return vars.FromPtr(&i) }
func newBoolVar(b bool) vars.PtrVar          { return vars.FromPtr(&b) }
func newListVar(l vals.List) vars.PtrVar     { return vars.FromPtr(&l) }
func newBindingVar(b bindingMap) vars.PtrVar { return vars.FromPtr(&b) }
