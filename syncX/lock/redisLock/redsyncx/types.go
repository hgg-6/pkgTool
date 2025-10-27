package redsyncx

type RedSyncIn interface {
	Start() <-chan LockResult
	Stop()
	IsLocked() bool
}
