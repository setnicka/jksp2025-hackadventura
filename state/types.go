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
	Metal  resultMetal
	Hotel  resultHotel
	Satna  resultSatna
	Csp  resultCsp
}

// Result for hack page
type Result struct {
	Completed     bool
	CompletedTime time.Time
	Tries         int
	LastTry       time.Time
}

func (r *Result) AddTry() {
	r.Tries++
	r.LastTry = time.Now()
}

func (r *Result) SetCompleted() {
	r.Completed = true
	r.CompletedTime = time.Now()
}

// Results for different hack pages
type resultMetal struct{ Result }
type resultHotel struct{ Result }
type resultSatna struct{ Result }
type resultCsp struct{ Result }
type resultTechno struct{ Result }

// when extra fields needed:
//
//	type ResultMoria struct {
//		result
//		RightAnswers int
//	}
