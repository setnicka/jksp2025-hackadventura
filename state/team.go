package state

import (
	"log/slog"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// GetTeam returns team by its login (or nil if not found)
func (s *State) GetTeam(login string) *Team {
	for i, team := range s.Teams {
		if team.Login == login {
			return s.Teams[i]
		}
	}
	return nil
}

// AddTeam creates a new team with given login and name (and saves the state)
func (s *State) AddTeam(login string, name string) error {
	if s.GetTeam(login) != nil {
		return errors.Errorf("Team with name '%s' already exists", login)
	}
	s.Teams = append(s.Teams, &Team{Login: login, Name: name})
	return s.Save()
}

// DeleteTeam removes team with given login (and saves the state)
func (s *State) DeleteTeam(login string) error {
	for i, team := range s.Teams {
		if team.Login == login {
			s.Teams = append(s.Teams[:i], s.Teams[i+1:]...)
			return s.Save()
		}
	}
	return errors.Errorf("Cannot find team with login '%s'", login)
}

// TeamSetPassword changes password for given team
func (s *State) TeamSetPassword(login string, password string) error {
	team := s.GetTeam(login)
	if team == nil {
		return errors.Errorf("Team '%s' not found", login)
	}
	slog.Debug("saving new password", "team", login)
	passwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	team.Passwd = passwd
	return s.Save()

}

// TeamLogin tries to find a team by login and check its password
func (s *State) TeamLogin(login string, password string) (bool, *Team) {
	team := s.GetTeam(login)
	if team == nil {
		return false, nil
	}
	if err := bcrypt.CompareHashAndPassword(team.Passwd, []byte(password)); err == nil {
		return true, team
	}
	return false, nil
}
