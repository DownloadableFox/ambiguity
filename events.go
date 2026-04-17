package ambiguity

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type EventContext[T any] struct {
	context.Context
	Session *discordgo.Session
	Event   *T
}

func NewEventContext[T any](session *discordgo.Session, event *T) *EventContext[T] {
	return &EventContext[T]{
		Context: context.Background(),
		Session: session,
		Event:   event,
	}
}

type EventExecuteFunc[T any] func(context *EventContext[T]) error
type EventMiddlewareFunc[T any] func(event Event[T], next EventExecuteFunc[T]) EventExecuteFunc[T]

type EventInfo struct {
	Name string
	Once bool
}

type Event[T any] interface {
	Info() EventInfo
	Handle(context *EventContext[T]) error
}

type EventMiddleware[T any] interface {
	Handle(event Event[T], next EventExecuteFunc[T]) EventExecuteFunc[T]
}

type CompilableEvent interface {
	Compile(session *discordgo.Session, log *slog.Logger) any
}

type LinkedEvent[T any] struct {
	Event       Event[T]
	Middlewares []EventMiddleware[T]
}

func LinkEvent[T any](event Event[T], middlewares ...EventMiddleware[T]) LinkedEvent[T] {
	return LinkedEvent[T]{
		Event:       event,
		Middlewares: middlewares,
	}
}

func (l *LinkedEvent[T]) Info() EventInfo {
	return l.Event.Info()
}

func (l *LinkedEvent[T]) Handle() EventExecuteFunc[T] {
	next := func(ctx *EventContext[T]) error {
		return l.Event.Handle(ctx)
	}

	for i := len(l.Middlewares) - 1; i >= 0; i-- {
		mw := l.Middlewares[i]
		next = mw.Handle(l.Event, next)
	}

	return next
}

func (l *LinkedEvent[T]) Compile(session *discordgo.Session, log *slog.Logger) any {
	handle := l.Handle()

	return func(s *discordgo.Session, e *T) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("Recovered from panic in event execution", slog.Any("panic", rec))
			}
		}()

		context := NewEventContext[T](session, e)
		if err := handle(context); err != nil {
			log.Error("Unhandled error in event execution", slog.Any("error", err))
		}
	}
}
