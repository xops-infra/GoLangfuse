// Package langfuse provides an interface to send ingestion events to Langfuse in an asynchronous manner.
package langfuse

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/bdpiprava/GoLangfuse/config"
	"github.com/bdpiprava/GoLangfuse/logger"
	"github.com/bdpiprava/GoLangfuse/types"
)

// maxParallelItem is the maximum number of items that can be processed in parallel.
const maxParallelItem = 512

// Langfuse an interface to send ingestion events to langfuse in async manner
// Event is added to the queue and then processor is sending it to the langfuse
type Langfuse interface {
	// AddEvent adds event to the channel and returns the event unique ID, generating one if missing.
	AddEvent(ctx context.Context, event types.LangfuseEvent) *uuid.UUID
}

type eventChanItem struct {
	ctx   context.Context
	event types.LangfuseEvent
}

type langfuseService struct {
	client       Client
	eventChannel chan eventChanItem
}

// New initialise new Langfuse instance for given config with background event processors
func New(config *config.Langfuse) Langfuse {
	if err := config.Validate(); err != nil {
		logger := logger.FromContext(context.Background())
		logger.Fatalf("invalid langfuse configuration: %v", err)
	}
	
	optimizedClient := NewOptimizedHTTPClient(config)
	return NewWithClient(config, optimizedClient)
}

// NewWithClient initialise new Langfuse instance with background event processors
func NewWithClient(config *config.Langfuse, customHTTPClient *http.Client) Langfuse {
	if err := config.Validate(); err != nil {
		logger := logger.FromContext(context.Background())
		logger.Fatalf("invalid langfuse configuration: %v", err)
	}
	
	eventManager := &langfuseService{
		client:       NewClient(config, customHTTPClient),
		eventChannel: make(chan eventChanItem, maxParallelItem),
	}
	eventManager.startEventProcessors(config.NumberOfEventProcessor)
	return eventManager
}

// AddEvent adds event to the channel and returns the event unique ID, generating one if missing.
func (l *langfuseService) AddEvent(ctx context.Context, event types.LangfuseEvent) *uuid.UUID {
	ensureEventID(event)
	l.eventChannel <- eventChanItem{ctx: ctx, event: event}
	return event.GetID()
}

// startEventProcessors start the background event processors
func (l *langfuseService) startEventProcessors(count int) {
	if count <= 0 {
		logrus.New().Warn("Langfuse event processor count is less than or equal to zero, no processors will be started")
	}

	for range count {
		go func() {
			for item := range l.eventChannel {
				l.send(item.ctx, item.event)
			}
		}()
	}
}

// send sends a types.LangfuseEvent to the Langfuse and logs any issues.
func (l *langfuseService) send(ctx context.Context, event types.LangfuseEvent) {
	log := logger.FromContext(ctx)
	log.Debugf("sending event to langfuse %v", event)

	err := l.client.Send(ctx, event)
	if err != nil {
		log.WithError(err).Errorf("failed to send event %v", event)
		return
	}
}

// ensureEventID ensures that the IngestionEvent has a unique ID, generating one if missing.
func ensureEventID(ingestionEvent types.LangfuseEvent) {
	if ingestionEvent.GetID() != nil {
		return
	}
	newID := uuid.New()
	ingestionEvent.SetID(&newID)
}
