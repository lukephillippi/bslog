package bslog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"testing/slogtest"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func TestHandler(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(
		slog.NewJSONHandler(&buf, nil),
		setupTestCollection(t),
	)

	results := func() []map[string]any {
		var ms []map[string]any
		for line := range bytes.SplitSeq(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	}

	if err := slogtest.TestHandler(h, results); err != nil {
		t.Fatal(err)
	}
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct{ attr slog.Attr }{
		{attr: slog.Any("any", struct{}{})},
		{attr: slog.Bool("bool", true)},
		{attr: slog.Duration("duration", time.Second)},
		{attr: slog.Float64("float64", 1.23)},
		{attr: slog.Int64("int64", 123)},
		{attr: slog.String("string", "string")},
		{attr: slog.Time("time", time.Now())},
		{attr: slog.Uint64("uint64", 123)},
	}

	for _, tt := range tests {
		t.Run(tt.attr.Key, func(t *testing.T) {
			var buf bytes.Buffer
			h := NewHandler(
				slog.NewJSONHandler(&buf, nil),
				setupTestCollection(t),
			)

			r := slog.Record{
				Time:    time.Now(),
				Level:   slog.LevelInfo,
				Message: "message",
			}
			r.AddAttrs(tt.attr)

			if err := h.Handle(context.Background(), r); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHandler_Handle_error(t *testing.T) {
	h := NewHandler(
		new(mockErrorHandler),
		setupTestCollection(t),
	)

	if err := h.Handle(context.Background(), slog.Record{}); err == nil {
		t.Errorf("expected error")
	}
}

func setupTestCollection(t *testing.T) *mongo.Collection {
	t.Helper()

	md := drivertest.NewMockDeployment()
	md.AddResponses(bson.D{{Key: "ok", Value: 1}})

	opts := options.Client()
	opts.Deployment = md

	client, err := mongo.Connect(opts)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Logf("Failed to disconnect client: %v", err)
		}
	})

	return client.Database("test").Collection("logs")
}

type mockErrorHandler struct{}

func (m *mockErrorHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (m *mockErrorHandler) Handle(context.Context, slog.Record) error { return fmt.Errorf("oops") }
func (m *mockErrorHandler) WithAttrs([]slog.Attr) slog.Handler        { return m }
func (m *mockErrorHandler) WithGroup(string) slog.Handler             { return m }
