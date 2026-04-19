// Package events is the platform's domain event bus, backed by NATS JetStream.
//
// Services publish and subscribe to well-known subjects like "bet.placed",
// "bet.settled", "wallet.debited" so that settlement, analytics, and
// fraud-detection can react asynchronously without coupling to the producer.
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Subject names. Add new ones here rather than sprinkling strings around.
const (
	SubjectBetPlaced     = "bet.placed"
	SubjectBetSettled    = "bet.settled"
	SubjectWalletCredit  = "wallet.credit"
	SubjectWalletDebit   = "wallet.debit"
	SubjectDepositOK     = "deposit.completed"
	SubjectWithdrawalOK  = "withdrawal.completed"
	SubjectGameRoundEnd  = "game.round.ended"
	SubjectKYCCompleted  = "kyc.completed"
)

// Envelope is the shape wrapping every published event. It allows consumers
// to route and version-evolve payloads without re-decoding twice.
type Envelope struct {
	ID        string          `json:"id"`
	Subject   string          `json:"subject"`
	OccurredAt time.Time      `json:"occurred_at"`
	Producer  string          `json:"producer"`
	Payload   json.RawMessage `json:"payload"`
}

// Bus is the publisher/subscriber interface our services depend on. The NATS
// implementation is in nats.go; tests can swap a no-op implementation here.
type Bus interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(subject string, handler func(context.Context, Envelope) error) (Unsubscribe, error)
	Close() error
}

// Unsubscribe cancels a previous Subscribe call.
type Unsubscribe func() error

// NATSBus is the NATS-backed Bus. It publishes plain NATS messages by default,
// but can be switched to JetStream for durability via WithJetStream.
type NATSBus struct {
	conn     *nats.Conn
	producer string
	js       nats.JetStreamContext
}

// Option configures a NATSBus.
type Option func(*NATSBus) error

// WithJetStream switches publishing and subscription to JetStream for
// at-least-once delivery with durable consumers.
func WithJetStream(cfg ...nats.JSOpt) Option {
	return func(b *NATSBus) error {
		js, err := b.conn.JetStream(cfg...)
		if err != nil {
			return fmt.Errorf("jetstream: %w", err)
		}
		b.js = js
		return nil
	}
}

// Connect establishes a connection to the NATS URL (e.g. "nats://localhost:4222")
// and returns a ready bus. The producer name is attached to every envelope.
func Connect(url, producer string, opts ...Option) (*NATSBus, error) {
	if url == "" {
		return nil, errors.New("nats url is required")
	}
	conn, err := nats.Connect(url,
		nats.Name(producer),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	b := &NATSBus{conn: conn, producer: producer}
	for _, opt := range opts {
		if err := opt(b); err != nil {
			_ = conn.Drain()
			return nil, err
		}
	}
	return b, nil
}

// Publish wraps payload in an Envelope and publishes to subject.
func (b *NATSBus) Publish(ctx context.Context, subject string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	env := Envelope{
		ID:         nats.NewInbox(),
		Subject:    subject,
		OccurredAt: time.Now().UTC(),
		Producer:   b.producer,
		Payload:    raw,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	if b.js != nil {
		_, err = b.js.PublishAsync(subject, data)
		return err
	}
	return b.conn.Publish(subject, data)
}

// Subscribe delivers decoded envelopes to handler. Handler errors are logged
// by the caller; NATS core does not retry (swap to JetStream if you need that).
func (b *NATSBus) Subscribe(subject string, handler func(context.Context, Envelope) error) (Unsubscribe, error) {
	wrap := func(msg *nats.Msg) {
		var env Envelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			return
		}
		_ = handler(context.Background(), env)
	}

	var (
		sub *nats.Subscription
		err error
	)
	if b.js != nil {
		sub, err = b.js.Subscribe(subject, wrap)
	} else {
		sub, err = b.conn.Subscribe(subject, wrap)
	}
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}
	return func() error { return sub.Unsubscribe() }, nil
}

// Close drains the connection and waits for in-flight publishes.
func (b *NATSBus) Close() error {
	if b.conn == nil {
		return nil
	}
	return b.conn.Drain()
}
