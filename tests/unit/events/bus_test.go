package events_test

import (
	"context"
	"errors"
	"testing"

	"pos-api/internal/pkg/events"
)

func TestEventBus_PublishAndSubscribe(t *testing.T) {
	bus := events.NewMemoryEventBus()
	ctx := context.Background()

	var called bool
	bus.Subscribe("test.event", func(ctx context.Context, payload interface{}) error {
		called = true
		if p, ok := payload.(string); ok && p == "hello" {
			return nil
		}
		return errors.New("invalid payload")
	})

	err := bus.Publish(ctx, "test.event", "hello")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestEventBus_ErrorPropagation(t *testing.T) {
	bus := events.NewMemoryEventBus()
	ctx := context.Background()

	expectedErr := errors.New("handler failed")

	bus.Subscribe("test.error", func(ctx context.Context, payload interface{}) error {
		return expectedErr
	})

	err := bus.Publish(ctx, "test.error", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) && err.Error() != "event handler for 'test.error' failed: handler failed" {
		t.Fatalf("expected specific error, got: %v", err)
	}
}
