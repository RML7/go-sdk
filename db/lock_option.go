package db

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type lockMode int

const (
	forUpdate lockMode = iota
	forNoKeyUpdate
	forShare
	forKeyShare
)

// WaitOption represents the lock wait strategy
type WaitOption int

const (
	// WaitOptionWait waits for the lock to be released (default)
	WaitOptionWait WaitOption = WaitOption(goqu.Wait)
	// WaitOptionNoWait returns an error immediately if the lock cannot be acquired
	WaitOptionNoWait WaitOption = WaitOption(goqu.NoWait)
	// WaitOptionSkipLocked skips locked rows
	WaitOptionSkipLocked WaitOption = WaitOption(goqu.SkipLocked)
)

// LockMode represents a row-level lock mode
type LockMode struct {
	lockMode   lockMode
	waitOption WaitOption
}

// ForUpdate returns a LockMode with FOR UPDATE lock (default: Wait)
func ForUpdate() LockMode {
	return LockMode{lockMode: forUpdate, waitOption: WaitOptionWait}
}

// ForNoKeyUpdate returns a LockMode with FOR NO KEY UPDATE lock (default: Wait)
func ForNoKeyUpdate() LockMode {
	return LockMode{lockMode: forNoKeyUpdate, waitOption: WaitOptionWait}
}

// ForShare returns a LockMode with FOR SHARE lock (default: Wait)
func ForShare() LockMode {
	return LockMode{lockMode: forShare, waitOption: WaitOptionWait}
}

// ForKeyShare returns a LockMode with FOR KEY SHARE lock (default: Wait)
func ForKeyShare() LockMode {
	return LockMode{lockMode: forKeyShare, waitOption: WaitOptionWait}
}

// Wait sets the lock to wait for release (default behavior)
func (m LockMode) Wait() LockMode {
	m.waitOption = WaitOptionWait
	return m
}

// NoWait sets the lock to return an error immediately if cannot be acquired
func (m LockMode) NoWait() LockMode {
	m.waitOption = WaitOptionNoWait
	return m
}

// SkipLocked sets the lock to skip locked rows
func (m LockMode) SkipLocked() LockMode {
	m.waitOption = WaitOptionSkipLocked
	return m
}

// WithLockMode adds a lock mode to the query
func WithLockMode(mode LockMode) ReadOption {
	return func(query *goqu.SelectDataset) *goqu.SelectDataset {
		switch mode.lockMode {
		case forUpdate:
			return query.ForUpdate(exp.WaitOption(mode.waitOption))
		case forNoKeyUpdate:
			return query.ForNoKeyUpdate(exp.WaitOption(mode.waitOption))
		case forShare:
			return query.ForShare(exp.WaitOption(mode.waitOption))
		case forKeyShare:
			return query.ForKeyShare(exp.WaitOption(mode.waitOption))
		default:
			return query
		}
	}
}
