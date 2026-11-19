package cli

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func newTestCLI() (*TCLI, *[][]int, *[][]float64, *[][]string) {
	var ints [][]int
	var floats [][]float64
	var strs [][]string
	c := &TCLI{}
	c.AddCommand("one", 1, 1, func(a []int) { ints = append(ints, a) }, "")
	c.AddCommand("four", 4, 4, func(a []float64) { floats = append(floats, a) }, "")
	c.AddCommand("range", 1, 2, func(a []string) { strs = append(strs, a) }, "")
	c.AddCommand("any", 0, Unlimited, func(a []string) { strs = append(strs, a) }, "")
	return c, &ints, &floats, &strs
}

func TestDispatchValidArguments(t *testing.T) {
	c, ints, floats, strs := newTestCLI()
	c.Dispatch("one 42")
	c.Dispatch("four 1 2.5 -3 4")
	c.Dispatch("range a b")
	if !reflect.DeepEqual(*ints, [][]int{{42}}) {
		t.Errorf("ints = %v", *ints)
	}
	if !reflect.DeepEqual(*floats, [][]float64{{1, 2.5, -3, 4}}) {
		t.Errorf("floats = %v", *floats)
	}
	if !reflect.DeepEqual(*strs, [][]string{{"a", "b"}}) {
		t.Errorf("strs = %v", *strs)
	}
}

func TestDispatchWrongArgumentCountDoesNotCallBack(t *testing.T) {
	c, ints, floats, strs := newTestCLI()
	for _, line := range []string{"one", "one 1 2", "four -2 2 -2", "four 1 2 3 4 5", "range", "range a b c"} {
		c.Dispatch(line)
	}
	if len(*ints)+len(*floats)+len(*strs) != 0 {
		t.Errorf("callbacks ran: %v %v %v", *ints, *floats, *strs)
	}
}

func TestDispatchUnparseableTokenRejectsWholeCommand(t *testing.T) {
	c, ints, floats, _ := newTestCLI()
	c.Dispatch("one 1x")
	c.Dispatch("four -2 2 oops 2") // must not shift later args into earlier slots
	if len(*ints)+len(*floats) != 0 {
		t.Errorf("callbacks ran: %v %v", *ints, *floats)
	}
}

func TestDispatchUnlimitedAndExtraSpaces(t *testing.T) {
	c, _, _, strs := newTestCLI()
	c.Dispatch("any")
	c.Dispatch("any a  b c ")
	want := [][]string{{}, {"a", "b", "c"}}
	if !reflect.DeepEqual(*strs, want) {
		t.Errorf("strs = %v, want %v", *strs, want)
	}
}

// captureStdout returns what f printed
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	f()
	os.Stdout = old
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestDispatchUnknownCommandIsReported(t *testing.T) {
	c, ints, _, _ := newTestCLI()
	for _, line := range []string{"iteration 1000", "One 5", "witdh 500"} {
		var handled bool
		out := captureStdout(t, func() { handled = c.Dispatch(line) })
		if handled {
			t.Errorf("%q reported as handled", line)
		}
		if !strings.Contains(out, "unknown command") || !strings.Contains(out, "help") {
			t.Errorf("%q: output %q lacks unknown-command message", line, out)
		}
	}
	if len(*ints) != 0 {
		t.Errorf("a callback ran: %v", *ints)
	}
}

func TestDispatchBlankLineIsQuiet(t *testing.T) {
	c, _, _, _ := newTestCLI()
	for _, line := range []string{"", "   "} {
		var handled bool
		out := captureStdout(t, func() { handled = c.Dispatch(line) })
		if out != "" {
			t.Errorf("%q printed %q", line, out)
		}
		if !handled {
			t.Errorf("%q returned false", line)
		}
	}
}

func TestDispatchKnownCommandIsHandledEvenWithBadArguments(t *testing.T) {
	c, _, _, _ := newTestCLI()
	var handled bool
	out := captureStdout(t, func() { handled = c.Dispatch("one") })
	if !handled || strings.Contains(out, "unknown command") {
		t.Errorf("handled=%v output=%q", handled, out)
	}
}

func TestDispatchStopsAtFirstMatch(t *testing.T) {
	var calls []string
	c := &TCLI{}
	c.AddCommand("dup", 0, 0, func(a []string) { calls = append(calls, "first") }, "")
	c.AddCommand("dup", 0, 0, func(a []string) { calls = append(calls, "second") }, "")
	c.Dispatch("dup")
	if !reflect.DeepEqual(calls, []string{"first"}) {
		t.Errorf("calls = %v, want only the first match", calls)
	}
}

func TestDispatchTrailingJunkIsRefused(t *testing.T) {
	c := &TCLI{}
	ran := false
	c.AddCommand("go", 0, 0, func(a []string) { ran = true }, "")
	captureStdout(t, func() { c.Dispatch("go 42") })
	if ran {
		t.Error("go ran with trailing arguments")
	}
}
