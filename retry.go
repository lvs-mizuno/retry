package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
)

var (
	flagInitialInterval = flag.Int("initialInterval", 1, "retry interval(s)")
	flagMaxElapsedTime  = flag.Int("maxElapsedTime", 10000, "Max Elapsed Time(s) is limit of backoff steps. If the job spends over this, job makes stopped. If set 0, the job will never stop.")
	flagMaxInterval     = flag.Int("maxInterval", 1000, "cap of retry interval(s)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: retry <command>\n\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
	}

	var b []byte
	operation := func() error {
		var err error
		b, err = exec.Command(flag.Arg(0), args[1:]...).Output()
		if err != nil {
			log.Printf("err: %s", err)
		}
		return err
	}

	bf := backoff.NewExponentialBackOff()
	second := func(i int) time.Duration {
		return time.Duration(i) * time.Second
	}

	bf.MaxElapsedTime = second(*flagMaxElapsedTime)
	bf.MaxInterval = second(*flagMaxInterval)
	bf.InitialInterval = second(*flagInitialInterval)

	err := backoff.Retry(operation, bf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "operation failed: %s\n", err)
		var exitStatus int
		if e, ok := err.(*exec.ExitError); ok {
			if s, ok := e.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			} else {
				panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
			}
		}
		fmt.Fprintf(os.Stdout, "\nlast stdout/stderr output:\n%s\n", string(b))
		os.Exit(exitStatus)
	}

	os.Exit(0)
}
