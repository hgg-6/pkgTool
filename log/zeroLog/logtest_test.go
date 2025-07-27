package main

import (
	"errors"
	"testing"
)

func TestInitLog(t *testing.T) {
	l := InitLogServer()
	l.Info().Err(errors.New("test error")).Int64("uid", 555).Msg("hello")
}
