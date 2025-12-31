package redsyncx

import "github.com/go-redsync/redsync/v4"

type RedSyncIn interface {
	Start() <-chan LockResult
	Stop()
	IsLocked() bool
	Status() LockStatus
	GetLockInfo() map[string]interface{}
	CreateMutex(name string) *redsync.Mutex
}
