package ambiguity

import "github.com/bwmarrin/discordgo"

func defaultAmbiguity(client *discordgo.Session) *Ambiguity {
	return &Ambiguity{
		client:          client,
		modules:         make([]Module, 0),
		modulesManager:  NewModuleManager(),
		eventsManager:   NewEventManager(),
		commandsManager: NewCommandManager(),
		tasksManager:    NewTaskManager(),
	}
}

func WithIntent(intent discordgo.Intent) AmbiguityOption {
	return func(a *Ambiguity) {
		a.client.Identify.Intents = intent
	}
}

func WithModules(modules ...Module) AmbiguityOption {
	return func(a *Ambiguity) {
		a.modules = append(a.modules, modules...)
	}
}

func WithSession(apply func(*discordgo.Session)) AmbiguityOption {
	return func(a *Ambiguity) {
		apply(a.client)
	}
}
