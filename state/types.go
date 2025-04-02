package state

import (
	"sync"
	"time"
)

// State of the game
type State struct {
	Teams  []*Team
	Global globalState
	sync.RWMutex
	Version int // state version, increased by each save
}

type globalState struct {
	Gundabad []GundabadPassword
}

type GundabadPassword struct {
	Password string
	From     time.Time
}

// Team holds teams login information and game state for it
type Team struct {
	Name   string
	Login  string
	Passwd []byte // hashed by bcrypt
}

// result for hack page
type result struct {
	Completed     bool
	CompletedTime time.Time
	Tries         int
	LastTry       time.Time
}

func (r *result) AddTry() {
	r.Tries++
	r.LastTry = time.Now()
}

func (r *result) SetCompleted() {
	r.Completed = true
	r.CompletedTime = time.Now()
}
