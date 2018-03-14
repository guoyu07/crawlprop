package main

import (
	"context"
	"fmt"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	devt := devtool.New("http://localhost:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		return
	}

	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close() // Cleanup.

	c := cdp.NewClient(conn)

	abort := make(chan error, 2)

	go func() {
		select {
		case <-ctx.Done():
		case err := <-abort:
			fmt.Printf("aborted: %s\n", err.Error())
			cancel()
		}
	}()

	if err = abortOnErrors(ctx, c, abort); err != nil {
		fmt.Println(err)
		return
	}

}

func abortOnErrors(ctx context.Context, c *cdp.Client, abort chan<- error) error {
	exceptionThrown, err := c.Runtime.ExceptionThrown(ctx)
	if err != nil {
		return err
	}

	loadingFailed, err := c.Network.LoadingFailed(ctx)
	if err != nil {
		return err
	}

	go func() {
		defer exceptionThrown.Close() // Cleanup.
		defer loadingFailed.Close()
		for {
			select {
			// Check for exceptions so we can abort as soon
			// as one is encountered.
			case <-exceptionThrown.Ready():
				ev, err := exceptionThrown.Recv()
				if err != nil {
					// This could be any one of: stream closed,
					// connection closed, context deadline or
					// unmarshal failed.
					abort <- err
					return
				}

				// Ruh-roh! Let the caller know something went wrong.
				abort <- ev.ExceptionDetails

			// Check for non-canceled resources that failed
			// to load.
			case <-loadingFailed.Ready():
				ev, err := loadingFailed.Recv()
				if err != nil {
					abort <- err
					return
				}

				// For now, most optional fields are pointers
				// and must be checked for nil.
				canceled := ev.Canceled != nil && *ev.Canceled

				if !canceled {
					abort <- fmt.Errorf("request %s failed: %s", ev.RequestID, ev.ErrorText)
				}
			}
		}
	}()
	return nil
}
