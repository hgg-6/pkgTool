package zerologx

import "github.com/rs/zerolog"

// Zlogger  接口（依赖抽象）
//
//go:generate mockgen -source=./types.go -package=zerologxmocks -destination=mocks/zerologx.mock.go Zlogger
type Zlogger interface {
	Info() *zerolog.Event
	Error() *zerolog.Event
	Debug() *zerolog.Event
	Warn() *zerolog.Event
	With() zerolog.Context
	GetZerolog() *zerolog.Logger
}
