// Example program demonstrating two ways to use spinner package
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/artyom/spinner"
)

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run() error {
	s := spinner.New(os.Stdout, "manual spinner...")
	for i := 0; i < 10; i++ {
		s.Spin()
		time.Sleep(time.Second / 10)
	}
	s.Clear()
	fmt.Println("done with stdout")
	foo()
	return nil
}

func foo() {
	defer fmt.Fprintln(os.Stderr, "done with stderr")
	defer spinner.Spin(os.Stderr, "helper function...")()
	time.Sleep(time.Second)
}
