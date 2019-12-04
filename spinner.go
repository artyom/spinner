// Package spinner implements simple "progress spinner" for terminal output.
//
// Spinner expects to have exclusive access to underlying *os.File, that
// nothing else is writing there while spinner is in use, otherwise output
// would be broken.
//
// If provided *os.File (usually os.Stdout or os.Sterr) is not attached to a
// terminal, spinner outputs nothing, that makes it safe to redirect program
// output to files, pipes, etc.
//
// Spinner can either be used manually, by first creating it with New function,
// then periodically calling Spin() method on it to refresh output and finally
// finishing with Clear() method call to clean output; or package-level Spin
// shortcut function can be used to launch background goroutine that handles
// output refresh.
package spinner

import (
	"os"
	"time"

	"golang.org/x/term"
)

// Spin is a shortcut function which creates new Spinner, launches background
// goroutine that periodically calls spinner's Spin method, and returns
// function that stops that background goroutine and clears spinner output.
//
// Can be used like this:
//
//  func work() {
//      defer spinner.Spin(os.Stderr, "working...")()
//      // logic here
//  }
//
// It is expected that nothing else is writing to underlying *os.File until
// stop function returns.
func Spin(f *os.File, text string) (stop func()) {
	s := New(f, text)
	if s.f == nil {
		return func() {}
	}
	done := make(chan struct{})
	bgDone := make(chan struct{})
	ticker := time.NewTicker(time.Second / 10)
	go func() {
		defer ticker.Stop()
		defer close(bgDone)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				s.Spin()
			}
		}
	}()
	return func() {
		close(done)
		<-bgDone
		s.Clear()
	}
}

// Spinner implements terminal spinner attached to *os.File which usually
// either stdout or stderr. Both zero and nil values are valid and are no-op.
// If spinner created on an *os.File that is not attached to the terminal,
// spinner's methods do nothing.
//
// Its methods are NOT thread safe, and it expects to have exclusive access to
// underlying *os.File — that nothing is writing to it while Spinner's methods
// are in use.
type Spinner struct {
	f    *os.File
	text []byte
	n    int
}

// Spin redraws output if underlying *os.File is attached to a terminal.
func (s *Spinner) Spin() {
	if s == nil || s.f == nil {
		return
	}
	const chars = `|/-\`
	s.n = (s.n + 1) % len(chars)
	s.text[len(s.text)-2] = chars[s.n]
	s.f.Write(s.text)
}

// Clear redraws output with spaces, clearing previous output if underlying
// *os.File is attached to a terminal.
func (s *Spinner) Clear() {
	if s == nil || s.f == nil {
		return
	}
	b := make([]byte, len(s.text))
	for i := range b {
		b[i] = ' '
	}
	b[len(b)-1] = '\r'
	s.f.Write(b)
}

// New returns new Spinner attached to f which usually either os.Stdout or
// os.Stderr. If f is attached to a terminal, retrurned spinner would output
// text followed by space and "spinning" character on each Spin call.
//
// *os.File provided must not be nil.
func New(f *os.File, text string) *Spinner {
	if !term.IsTerminal(int(f.Fd())) {
		return &Spinner{}
	}

	// b is "text"+" "+spinchar+"\r"
	b := make([]byte, len(text)+3)
	copy(b, []byte(text))
	b[len(b)-3] = ' '
	// b[len(b)-2] is replaced on each Spin() call
	b[len(b)-1] = '\r'
	return &Spinner{f: f, text: b}
}
