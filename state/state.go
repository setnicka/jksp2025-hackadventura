package state

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
)

const (
	logsDir       = "logs"
	stateFilename = "state.json"
)

// Init loads (or creates a new) game state and returns it. It is a caller
// responsibility to correctly use the state.Lock() and state.Rlock()
func Init() *State {
	slog.Debug("initializing game state")
	state := &State{
		Teams: []*Team{},
	}

	slog := slog.With("file", stateFilename)

	// Try to load previously saved state
	jsonFile, err := os.Open(stateFilename)
	if err == nil {
		defer jsonFile.Close()

		slog.Debug("loading state from file")
		jsonBytes, _ := io.ReadAll(jsonFile)
		if err = json.Unmarshal(jsonBytes, &state); err != nil {
			slog.Error("problem during loading state from file", "err", err)
		} else {
			slog.Debug("game state loaded")
		}
	}

	return state
}

// Save current state to the file
func (s *State) Save() error {
	slog := slog.With("file", stateFilename)

	slog.Debug("saving actual state into file")
	s.Version++
	// 1. If exists current state move it into folder
	if _, err := os.Stat(stateFilename); err == nil {
		if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "Cannot create directory '%s' for logging old states", logsDir)
		}
		filename := fmt.Sprintf("%s%s", stateFilename, time.Now().Format(".150405.00")) // 2006-01-02_150405
		if err := os.Rename(stateFilename, path.Join(logsDir, filename)); err != nil {
			return errors.Wrapf(err, "Cannot move state to logs directory '%s", logsDir)
		}
	}

	// 2. Marshal state into json
	bytes, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return errors.Wrapf(err, "Cannot save actual state into json")
	}

	if err := os.WriteFile(stateFilename, bytes, 0644); err != nil {
		return errors.Wrapf(err, "Cannot save JSON of the actual state into file '%s'", stateFilename)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// global state handlers

// SetGundabadPassword appends a new password for Dol Guldur (and save the state)
func (s *State) SetGundabadPassword(password string) error {
	s.Global.Gundabad = append(s.Global.Gundabad, GundabadPassword{Password: password, From: time.Now()})
	return s.Save()
}

// GetGundabadPasswords returns all passwords for Dol Guldur, valid one is the
// last one
func (s *State) GetGundabadPasswords() []GundabadPassword {
	return s.Global.Gundabad
}
