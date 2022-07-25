package keygroup

import (
	"context"
	"sync"

	"github.com/scylladb/go-set/strset"
)

// Func is the function signature to be passed to `NewKeyGroup`, it receives
// the context and a key string. If the function is running as a closed-loop,
// it should use `<-ctx.Done()` to know when to halt.
type Func func(context.Context, string)

// KeyGroup is a wrapper to `sync.WaitGroup` which allows you to coordinate
// goroutines based on string keys.
type KeyGroup struct {
	keys     *strset.Set
	cancel   map[string]context.CancelFunc
	function Func
	mainCtx  context.Context
	wg       sync.WaitGroup
}

// NewKeyGroup creates a `KeyGroup`
func NewKeyGroup(f Func) *KeyGroup {
	return NewKeyGroupWithContext(context.Background(), f)
}

// NewKeyGroupWithContext is same s `NewKeyGroup` but accepts a Context passed
// from the caller.
func NewKeyGroupWithContext(ctx context.Context, f Func) *KeyGroup {
	mainCtx, cancel := context.WithCancel(ctx)
	return &KeyGroup{
		keys: strset.New(),
		cancel: map[string]context.CancelFunc{
			"main": cancel,
		},
		mainCtx:  mainCtx,
		function: f,
		wg:       sync.WaitGroup{},
	}
}

// Cancel sends a stop signal to all running routines through their context.
func (kg *KeyGroup) Cancel() {
	kg.cancel["main"]()
}

func (kg *KeyGroup) run(keys []string) {
	for _, key := range keys {
		ctx, cancel := context.WithCancel(kg.mainCtx)
		kg.cancel[key] = cancel
		kg.wg.Add(1)
		localKey := key
		go func() {
			kg.function(ctx, localKey)
			kg.wg.Done()
			kg.keys.Remove(localKey)
		}()
	}
}

func (kg *KeyGroup) stop(keys []string) {
	for _, key := range keys {
		kg.cancel[key]()
		delete(kg.cancel, key)
	}
}

// Update receives a new list of key strings and determines which keys to stop
// and which keys to start.
func (kg *KeyGroup) Update(keys []string) {
	current := strset.New(keys...)
	toStart := strset.Difference(current, kg.keys)
	toStop := strset.Difference(kg.keys, current)
	kg.run(toStart.List())
	kg.stop(toStop.List())
	kg.keys = current
}

// Wait blocks until all routines are done.
func (kg *KeyGroup) Wait() {
	kg.wg.Wait()
}

// CancelWait cancels all running routines and wait for them to finish.
func (kg *KeyGroup) CancelWait() {
	kg.Cancel()
	kg.Wait()
}

// Add adds more keys to an already running `KeyGroup`
func (kg *KeyGroup) Add(keys ...string) {
	kg.Update(append(kg.keys.List(), keys...))
}

// StopKeys sends a termination signal to specific keys through their context.
func (kg *KeyGroup) StopKeys(keys ...string) {
	kg.stop(keys)
	kg.keys.Remove(keys...)
}
