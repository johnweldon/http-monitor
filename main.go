package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:generate go run gen.go

var (
	checkInterval = 30 * time.Second
	timeout       = 15 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	inch, errch := checkLoop(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(checkInterval):
				for _, c := range checks {
					inch <- c
				}
			}
		}
	}()

	go func() {
		for err := range errch {
			log.Printf("ERROR: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Printf("cancelling...\n")
	defer cancel()
}

func checkLoop(ctx context.Context) (chan<- check, <-chan error) {
	errChan := make(chan error)
	inChan := make(chan check)
	cl := http.Client{Timeout: timeout}

	go func() {

		var (
			err error
			req *http.Request
			res *http.Response
		)

		for {
			select {
			case c := <-inChan:
				if req, err = c.makeRequest(); err != nil {
					errChan <- &CheckError{cause: err, msg: "cannot make request", c: c}
					continue
				}
				if res, err = cl.Do(req); err != nil {
					errChan <- &CheckError{cause: err, msg: "cannot execute request", c: c}
					continue
				}
				if msg, ok := c.Validate(res); !ok {
					errChan <- &CheckError{msg: msg, c: c}
				}
			case <-ctx.Done():
				close(inChan)
				close(errChan)
				return
			}
		}
	}()

	return inChan, errChan
}

// CheckError wraps checker errors
type CheckError struct {
	cause error
	msg   string
	c     check
}

func (e *CheckError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("%s [%s %s]", e.msg, e.c.Method, e.c.URL)
	}
	return fmt.Sprintf("%s because %v", e.msg, e.cause)
}

func (e *CheckError) Unwrap() error { return e.cause }

// URL returns the url that caused this error
func (e *CheckError) URL() string { return e.c.URL }

// Method returns the method that caused this error
func (e *CheckError) Method() string { return e.c.Method }
