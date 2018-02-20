package tests

import (
	"testing"
	"time"
)

func TestSimpleSingle(t *testing.T) {
	time.Sleep(time.Second * 5)
	t.Log("TODO")
	t.Error("foo")
}
