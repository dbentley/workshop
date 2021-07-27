package wslocal

import (
	"context"
	"log"
	"time"

	"github.com/tilt-dev/workshop/wslocal/apihelpers"
	"github.com/tilt-dev/workshop/wslocal/first"
	"github.com/tilt-dev/workshop/wslocal/state"
)

type Machine interface {
	Advance(ctx context.Context, pre state.State) (state.State, error)
	PrintState(ctx context.Context, st state.State) error
}

type Looper struct {
	machine Machine
	api     *apihelpers.APIHelper
}

func NewLooper() (*Looper, error) {
	// here is where you would plug a different state machine

	api, err := apihelpers.NewAPIHelper()
	if err != nil {
		return nil, err
	}

	m, err := first.NewMachine(api)
	if err != nil {
		return nil, err
	}

	return &Looper{machine: m, api: api}, nil
}

func (l *Looper) Loop() error {
	ctx := context.Background()
	first := true
	for {
		if err := l.Iter(ctx, first); err != nil {
			log.Printf("error running workshop server: %v", err)
		}
		first = false
		time.Sleep(100 * time.Millisecond)
	}

	// unreachable
}

func (l *Looper) Iter(ctx context.Context, first bool) error {
	preSt, err := state.LoadState()
	if err != nil {
		return err
	}

	var postSt state.State

	interSt := preSt

	for {
		postSt, err = l.machine.Advance(ctx, state.CopyState(interSt))
		if err != nil {
			return err
		}
		if postSt.StateName == interSt.StateName {
			break
		}
		interSt = postSt
	}

	newState, err := state.SaveState(ctx, l.api, preSt, postSt, first)

	if err != nil {
		return err
	}

	if newState {
		return l.PrintState(ctx, postSt)
	}

	return nil
}

func (l *Looper) PrintState(ctx context.Context, st state.State) error {
	return l.machine.PrintState(ctx, st)
}
