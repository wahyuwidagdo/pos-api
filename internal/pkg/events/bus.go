package events

import (
	"context"
	"fmt"
)

// EventHandler is a function that processes an event.
// It returns an error if the processing fails, which will generally
// bubble up and abort the transaction/operation that triggered the event.
type EventHandler func(ctx context.Context, payload interface{}) error

// EventBus defines the interface for our internal Pub/Sub system.
type EventBus interface {
	// Subscribe registers a handler for a specific event name.
	Subscribe(eventName string, handler EventHandler)

	// Publish synchronously triggers all handlers registered for the given event name.
	// If any handler returns an error, Publish stops and returns that error immediately.
	Publish(ctx context.Context, eventName string, payload interface{}) error
}

type memoryEventBus struct {
	handlers map[string][]EventHandler
}

// NewMemoryEventBus creates a new synchronous, in-memory EventBus.
func NewMemoryEventBus() EventBus {
	return &memoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (b *memoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *memoryEventBus) Publish(ctx context.Context, eventName string, payload interface{}) error {
	handlers, ok := b.handlers[eventName]
	if !ok {
		// No handlers registered for this event; that's fine.
		return nil
	}

	for _, handler := range handlers {
		if err := handler(ctx, payload); err != nil {
			return fmt.Errorf("event handler for '%s' failed: %w", eventName, err)
		}
	}

	return nil
}
