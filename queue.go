package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueManager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	replies chan amqp.Delivery
}

type RPCRequest struct {
	Method    string          `json:"method"`
	Payload   json.RawMessage `json:"payload"`
	SessionID string          `json:"sessionId,omitempty"`
}

type RPCResponse struct {
	Success   bool            `json:"success"`
	SessionID string          `json:"sessionId,omitempty"`
	Response  map[string]any `json:"response,omitempty"`
	Type      string          `json:"type,omitempty"`
	Error     string          `json:"error,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

func NewQueueManager(url string) (*QueueManager, error) {
	var conn *amqp.Connection
	var err error

	// Retry connection
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying in 5s... (%d/5)", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Setup Dead Letter Exchange
	err = ch.ExchangeDeclare(
		"dlx_exchange", // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		"dlx_queue", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind("dlx_queue", "dlx_routing_key", "dlx_exchange", false, nil)
	if err != nil {
		return nil, err
	}

	// Setup main RPC queue with DLX
	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dlx_routing_key",
		"x-max-priority":            uint8(10),
	}

	_, err = ch.QueueDeclare(
		"rpc_queue", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		args,        // arguments
	)
	if err != nil {
		return nil, err
	}

	qm := &QueueManager{
		conn:    conn,
		channel: ch,
	}

	return qm, nil
}

func (qm *QueueManager) CallRPC(ctx context.Context, method string, sessionID string, payload interface{}, priority uint8, ttl time.Duration) (*RPCResponse, error) {
	corrId := fmt.Sprintf("%d", time.Now().UnixNano())

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req := RPCRequest{
		Method:    method,
		Payload:   rawPayload,
		SessionID: sessionID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	q, err := qm.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	msgs, err := qm.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}

	headers := amqp.Table{}
	if priority > 0 {
		headers["x-priority"] = int(priority)
	}

	expiration := ""
	if ttl > 0 {
		expiration = fmt.Sprintf("%d", ttl.Milliseconds())
	}

	err = qm.channel.PublishWithContext(ctx,
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          body,
			Priority:      priority,
			Expiration:    expiration,
			Headers:       headers,
		})
	if err != nil {
		return nil, err
	}

	for d := range msgs {
		if corrId == d.CorrelationId {
			var res RPCResponse
			err := json.Unmarshal(d.Body, &res)
			if err != nil {
				return nil, err
			}
			res.Raw = d.Body
			return &res, nil
		}
	}

	return nil, fmt.Errorf("no response from RPC")
}

func (qm *QueueManager) Close() {
	if qm.channel != nil {
		qm.channel.Close()
	}
	if qm.conn != nil {
		qm.conn.Close()
	}
}
