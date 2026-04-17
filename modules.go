package ambiguity

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Module interface{}

type ModuleWithCommands interface {
	Module
	Commands() ([]LinkedCommand, error)
}

type ModuleWithEvents interface {
	Module
	Events() ([]CompilableEvent, error)
}

type ModuleWithConfigure interface {
	Module
	Configure(session *discordgo.Session) error
}

type ModuleWithShutdown interface {
	Module
	Shutdown(session *discordgo.Session) error
}

func (a *Ambiguity) bootstrapModules() {
	a.log.Info("Loading modules...", slog.Int("count", len(a.modules)))

	var wg sync.WaitGroup
	for _, module := range a.modules {
		wg.Go(func() { a.loadModule(module) })
	}

	wg.Wait()
	a.log.Info("Done loading modules!")
}

func (a *Ambiguity) removeModules() {
	a.log.Info("Shutting down modules...")

	var wg sync.WaitGroup
	for _, module := range a.modules {
		wg.Go(func() { a.removeModule(module) })
	}

	wg.Wait()
	a.log.Info("Done loading modules!")
}

func (a *Ambiguity) loadModule(module Module) {
	// 1. Load modules.
	if err := a.configureModule(module); err != nil {
		a.log.Warn("Failed to configure module!",
			slog.String("module", fmt.Sprintf("%T", module)),
			slog.Any("error", err),
		)
	}

	// 2. Register commands.
	if err := a.registerCommands(module); err != nil {
		a.log.Warn("Failed to register commands for module!",
			slog.String("module", fmt.Sprintf("%T", module)),
			slog.Any("error", err),
		)
	}

	// 3. Publish commands.
	if err := a.publishCommands(module); err != nil {
		a.log.Warn("Failed to publish commands for module!",
			slog.String("module", fmt.Sprintf("%T", module)),
			slog.Any("error", err),
		)
	}

	// 4. Register events
	if err := a.registerEvents(module); err != nil {
		a.log.Warn("Failed to register events for module!",
			slog.String("module", fmt.Sprintf("%T", module)),
			slog.Any("error", err),
		)
	}

	// 5. Register tasks
}

func (a *Ambiguity) removeModule(module Module) {
	if err := a.shutdownModule(module); err != nil {
		a.log.Warn("Failed to shutdown module!",
			slog.String("module", fmt.Sprintf("%T", module)),
			slog.Any("error", err),
		)
	}
}

func (a *Ambiguity) configureModule(module Module) error {
	moduleWithConfigure, ok := module.(ModuleWithConfigure)
	if !ok {
		return nil
	}

	return moduleWithConfigure.Configure(a.session)
}

func (a *Ambiguity) shutdownModule(module Module) error {
	moduleWithShutdown, ok := module.(ModuleWithShutdown)
	if !ok {
		return nil
	}

	return moduleWithShutdown.Shutdown(a.session)
}

func (a *Ambiguity) registerCommands(module Module) error {
	moduleWithCommands, ok := module.(ModuleWithCommands)
	if !ok {
		return nil
	}

	commands, err := moduleWithCommands.Commands()
	if err != nil {
		return fmt.Errorf("obtain commands: %w", err)
	}

	for _, command := range commands {
		a.session.AddHandler(command.Compile(a.session, a.log))
	}

	return nil
}

func (a *Ambiguity) publishCommands(module Module) error {
	moduleWithCommands, ok := module.(ModuleWithCommands)
	if !ok {
		return nil
	}

	commands, err := moduleWithCommands.Commands()
	if err != nil {
		return fmt.Errorf("obtain commands: %w", err)
	}

	for _, command := range commands {
		info := command.Info()

		existing, err := a.commandStore.Load(info.Name)
		switch {
		case errors.Is(err, ErrCommandNotFoundInStore):
			created, err := a.session.ApplicationCommandCreate(a.session.State.User.ID, "", &info)
			if err != nil {
				return fmt.Errorf("command publish: %w", err)
			}

			if err := a.commandStore.Save(created); err != nil {
				return fmt.Errorf("command save: %w", err)
			}
		case err == nil:
			updated, err := a.session.ApplicationCommandEdit(a.session.State.User.ID, "", existing.ID, &info)
			if err != nil {
				return fmt.Errorf("command update: %w", err)
			}

			if err := a.commandStore.Save(updated); err != nil {
				return fmt.Errorf("command save: %w", err)
			}
		default:
			return err
		}
	}

	return nil
}

func (a *Ambiguity) registerEvents(module Module) error {
	moduleWithEvents, ok := module.(ModuleWithEvents)
	if !ok {
		return nil
	}

	events, err := moduleWithEvents.Events()
	if err != nil {
		return fmt.Errorf("obtain events: %w", err)
	}

	for _, event := range events {
		a.session.AddHandler(event.Compile(a.session, a.log))
	}

	return nil
}
