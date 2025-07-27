//go:build wireinject

package main

import (
	"github.com/google/wire"
)

func InitLogServer() Zlogger {
	wire.Build(
		InitLog,
	)
	return nil
}
