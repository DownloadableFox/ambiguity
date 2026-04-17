package ambiguity

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type AmbiguityOption func(a *Ambiguity)

type Ambiguity struct {
	log          *slog.Logger
	session      *discordgo.Session
	modules      []Module
	commandStore CommandStore
}

func NewWithSession(session *discordgo.Session, options ...AmbiguityOption) *Ambiguity {
	a := &Ambiguity{
		session:      session,
		modules:      make([]Module, 0),
		commandStore: &NoOpCommandStore{},
	}

	for _, apply := range options {
		apply(a)
	}

	a.bootstrapModules()
	return a
}

func (a *Ambiguity) Start() error {
	a.bootstrapModules()
	if err := a.session.Open(); err != nil {
		return fmt.Errorf("connect to discord: %w", err)
	}

	return nil
}

func (a *Ambiguity) Close() error {
	a.removeModules()
	return a.session.Close()
}
