package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/valkey-io/valkey-go"
)

type ValkeyClient struct {
	client valkey.Client
	logger *slog.Logger
}

func NewValkeyClient(addr string, logger *slog.Logger) (*ValkeyClient, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
	})
	if err != nil {
		return nil, fmt.Errorf("valkey connect: %w", err)
	}

	logger.Info("connected to valkey", "addr", addr)
	return &ValkeyClient{client: client, logger: logger}, nil
}

func (v *ValkeyClient) PushJob(ctx context.Context, queueKey string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return v.client.Do(ctx, v.client.B().Lpush().Key(queueKey).Element(string(data)).Build()).Error()
}

func (v *ValkeyClient) Publish(ctx context.Context, channel string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal publish: %w", err)
	}
	return v.client.Do(ctx, v.client.B().Publish().Channel(channel).Message(string(data)).Build()).Error()
}

func (v *ValkeyClient) Subscribe(ctx context.Context, channel string) (<-chan string, func()) {
	ch := make(chan string, 64)
	dedicated, cancel := v.client.Dedicate()

	go func() {
		defer close(ch)
		wait := dedicated.SetPubSubHooks(valkey.PubSubHooks{
			OnMessage: func(m valkey.PubSubMessage) {
				select {
				case ch <- m.Message:
				case <-ctx.Done():
				}
			},
		})
		_ = dedicated.Do(ctx, dedicated.B().Subscribe().Channel(channel).Build()).Error()
		<-wait
	}()

	return ch, cancel
}

func (v *ValkeyClient) Ping(ctx context.Context) error {
	return v.client.Do(ctx, v.client.B().Ping().Build()).Error()
}

func (v *ValkeyClient) Close() {
	v.client.Close()
}
