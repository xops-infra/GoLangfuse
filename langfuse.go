// Package langfuse provides an interface to send ingestion events to Langfuse in an asynchronous manner.
package langfuse

import (
	"context"
	"net/http"
	"sync"
	"time"

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
	// Stop gracefully shuts down the service and flushes remaining events
	Stop(ctx context.Context) error
}

type eventChanItem struct {
	ctx   context.Context
	event types.LangfuseEvent
}

type langfuseService struct {
	client       Client
	config       *config.Langfuse
	eventChannel chan eventChanItem
	stopChannel  chan struct{}
	wg           sync.WaitGroup
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
		config:       config,
		eventChannel: make(chan eventChanItem, maxParallelItem),
		stopChannel:  make(chan struct{}),
	}
	eventManager.startBatchProcessors(config.NumberOfEventProcessor)
	return eventManager
}

// AddEvent adds event to the channel and returns the event unique ID, generating one if missing.
func (l *langfuseService) AddEvent(ctx context.Context, event types.LangfuseEvent) *uuid.UUID {
	ensureEventID(event)
	l.eventChannel <- eventChanItem{ctx: ctx, event: event}
	return event.GetID()
}

// startBatchProcessors start the background batch processors
func (l *langfuseService) startBatchProcessors(count int) {
	if count <= 0 {
		logrus.New().Warn("Langfuse event processor count is less than or equal to zero, no processors will be started")
		return
	}

	for i := 0; i < count; i++ {
		l.wg.Add(1)
		go func(processorID int) {
			defer l.wg.Done()
			l.processBatches(processorID)
		}(i)
	}
}

// processBatches processes events in batches with timeout-based flushing
func (l *langfuseService) processBatches(processorID int) {
	log := logger.FromContext(context.Background())
	log.Debugf("Starting batch processor %d", processorID)

	var batch []eventChanItem
	ticker := time.NewTicker(l.config.BatchTimeout)
	defer ticker.Stop()

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		// Group events by context (for better tracing)
		contextGroups := make(map[context.Context][]types.LangfuseEvent)
		for _, item := range batch {
			contextGroups[item.ctx] = append(contextGroups[item.ctx], item.event)
		}

		// Send each context group as a batch
		for ctx, events := range contextGroups {
			l.sendBatch(ctx, events)
		}

		batch = batch[:0] // Clear the batch
	}

	for {
		select {
		case item, ok := <-l.eventChannel:
			if !ok {
				// Channel closed, flush remaining events and exit
				flushBatch()
				log.Debugf("Batch processor %d stopped", processorID)
				return
			}

			batch = append(batch, item)

			// Flush batch if it reaches the configured size
			if len(batch) >= l.config.BatchSize {
				flushBatch()
			}

		case <-ticker.C:
			// Flush batch on timeout
			flushBatch()

		case <-l.stopChannel:
			// Graceful shutdown requested
			flushBatch()
			log.Debugf("Batch processor %d stopped gracefully", processorID)
			return
		}
	}
}

// sendBatch sends a batch of events to Langfuse and logs any issues
func (l *langfuseService) sendBatch(ctx context.Context, events []types.LangfuseEvent) {
	log := logger.FromContext(ctx)
	log.Debugf("sending batch of %d events to langfuse", len(events))

	err := l.client.SendBatch(ctx, events)
	if err != nil {
		log.WithError(err).Errorf("failed to send batch of %d events", len(events))
		// Fall back to individual sends on batch failure
		for _, event := range events {
			if sendErr := l.client.Send(ctx, event); sendErr != nil {
				log.WithError(sendErr).Errorf("failed to send individual event %v", event)
			}
		}
	}
}

// Stop gracefully shuts down the service and flushes remaining events
func (l *langfuseService) Stop(ctx context.Context) error {
	log := logger.FromContext(ctx)
	log.Info("Stopping Langfuse service...")

	// Signal all processors to stop
	close(l.stopChannel)

	// Close the event channel to signal no more events
	close(l.eventChannel)

	// Wait for all processors to finish with timeout
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("Langfuse service stopped gracefully")
		return nil
	case <-ctx.Done():
		log.Warn("Langfuse service stop timed out")
		return ctx.Err()
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
