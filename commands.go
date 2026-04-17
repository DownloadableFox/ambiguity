package ambiguity

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type DiscordGoCommandHandler func(session *discordgo.Session, interaction *discordgo.InteractionCreate)
type CommandExecuteFunc func(context *CommandContext) error
type CommandMiddlewareFunc func(command Command, next CommandExecuteFunc) CommandExecuteFunc

type CommandContext struct {
	context.Context
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
}

func NewCommandContext(session *discordgo.Session, interaction *discordgo.InteractionCreate) *CommandContext {
	return &CommandContext{
		Context:     context.Background(),
		Session:     session,
		Interaction: interaction,
	}
}

func (c *CommandContext) Options() CommandOptions {
	return CommandOptions{
		options: c.Interaction.ApplicationCommandData().Options,
	}
}

type Command interface {
	Info() discordgo.ApplicationCommand
	Handle(context *CommandContext) error
}

type CommandMiddleware interface {
	Handle(command Command, next CommandExecuteFunc) CommandExecuteFunc
}

// A command alongside it's middlewares
type LinkedCommand struct {
	Command     Command
	Middlewares []CommandMiddleware
}

// Creates a command stack from the given
func LinkCommand(command Command, middleware ...CommandMiddleware) LinkedCommand {
	return LinkedCommand{
		Command:     command,
		Middlewares: middleware,
	}
}

func (l *LinkedCommand) Info() discordgo.ApplicationCommand {
	return l.Command.Info()
}

func (l *LinkedCommand) Handle() CommandExecuteFunc {
	// Register the command with the middleware
	next := func(ctx *CommandContext) error {
		return l.Command.Handle(ctx)
	}

	// Execute the middleware in reverse order
	// to ensure the first middleware is executed last
	for i := len(l.Middlewares) - 1; i >= 0; i-- {
		mw := l.Middlewares[i]
		next = mw.Handle(l.Command, next)
	}

	return next
}

func (l *LinkedCommand) Compile(session *discordgo.Session, log *slog.Logger) DiscordGoCommandHandler {
	handle := l.Handle()

	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if interaction.ApplicationCommandData().Name != l.Info().Name {
			return
		}

		defer func() {
			if rec := recover(); rec != nil {
				log.Error("Recovered from panic in command execution", slog.Any("panic", rec))
			}
		}()

		context := NewCommandContext(session, interaction)
		if err := handle(context); err != nil {
			log.Error("Unhandled error in command execution", slog.Any("error", err))
		}
	}
}
