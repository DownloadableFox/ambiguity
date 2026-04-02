package ambiguity

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Ambiguity struct {
	client  *discordgo.Session
	modules []Module

	modulesManager  ModuleManager
	eventsManager   EventManager
	commandsManager CommandManager
	tasksManager    TaskManager
}

type AmbiguityOption func(*Ambiguity)

func New(token string, options ...AmbiguityOption) (*Ambiguity, error) {
	client, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	client.StateEnabled = true
	client.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions

	a := defaultAmbiguity(client)
	for _, apply := range options {
		apply(a)
	}

	if err := a.modulesManager.RegisterModules(a.modules...); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Ambiguity) Start() error {
	if err := a.modulesManager.OnEvents(a.client, a.eventsManager); err != nil {
		return fmt.Errorf("register events: %w", err)
	}

	// Run the bot until terminated
	if err := a.client.Open(); err != nil {
		return fmt.Errorf("connect to discord: %w", err)
	}
	defer a.client.Close()

	// Register commands
	if err := a.modulesManager.OnCommands(a.client, a.commandsManager); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	// Register tasks
	if err := a.modulesManager.OnTasks(a.client, a.tasksManager); err != nil {
		return fmt.Errorf("register tasks: %w", err)
	}

	return nil
}
