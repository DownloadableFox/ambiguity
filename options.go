package ambiguity

import "github.com/bwmarrin/discordgo"

func WithModules(modules ...Module) AmbiguityOption {
	return func(a *Ambiguity) {
		a.modules = modules
	}
}

func WithCommandStore(store CommandStore) AmbiguityOption {
	return func(a *Ambiguity) {
		a.commandStore = store
	}
}

func WithIntents(intents discordgo.Intent) AmbiguityOption {
	return func(a *Ambiguity) {
		a.session.Identify.Intents = intents
	}
}
