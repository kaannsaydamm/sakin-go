package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"

	"sakin-go/cmd/sge-ingest/normalizer"
	"sakin-go/pkg/messaging"
)

type EventHandler struct {
	natsClient *messaging.Client
}

func NewEventHandler(nc *messaging.Client) *EventHandler {
	return &EventHandler{natsClient: nc}
}

// HandleHTTPEvent receives events via HTTP POST.
func (h *EventHandler) HandleHTTPEvent(c *fiber.Ctx) error {
	// 1. Get Raw Body (Zero Allocation in Fiber)
	body := c.Body()

	// 2. Normalize
	evt, err := normalizer.NormalizeAgentEvent(body)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid event format"})
	}

	// 3. Serialize for Bus
	// Can optimize by using same buffer if normalization supports it
	data, _ := json.Marshal(evt) // In real world use custom serializer

	// 4. Publish to NATS (Async)
	// Topic: events.raw.<severity>.<source>
	subject := messaging.TopicEventsRaw + string(evt.Severity) + "." + evt.Source

	_, err = h.natsClient.PublishAsync(context.Background(), subject, data)
	if err != nil {
		log.Printf("[Ingest] NATS Publish Error: %v", err)
		return c.Status(500).SendString("Internal Bus Error")
	}

	return c.SendStatus(202) // Accepted
}
