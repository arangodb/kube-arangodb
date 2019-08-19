package tests

import (
	"fmt"
	"time"
)

type interrupt struct {
}

func (i interrupt) Error() string {
	return "interrupted"
}

func isInterrupt(err error) bool {
	_, ok := err.(interrupt)
	return ok
}

func timeout(interval, timeout time.Duration, action func() error) error {
	intervalT := time.NewTicker(interval)
	defer intervalT.Stop()

	timeoutT := time.NewTimer(timeout)
	defer timeoutT.Stop()

	for {
		select {
		case <-intervalT.C:
			err := action()
			if err != nil {
				if isInterrupt(err) {
					return nil
				}
				return err
			}
			break
		case <-timeoutT.C:
			return fmt.Errorf("function timeouted")
		}
	}
}
