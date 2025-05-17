// Package bslog provides log archival as MongoDB BSON documents for [log/slog].
package bslog // import "go.luke.ph/bslog"

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// NewHandler constructs a [*Handler] that wraps the given handler.
func NewHandler(handler slog.Handler, collection *mongo.Collection) *Handler {
	return &Handler{
		handler:    handler,
		collection: collection,
	}
}

// A Handler implements the [slog.Handler] interface.
type Handler struct {
	handler    slog.Handler
	collection *mongo.Collection
}

// Enabled implements the [slog.Handler] Enabled interface method.
// It calls the wrapped handler's Enabled method directly.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements the [slog.Handler] Handle interface method.
// It calls the wrapped handler's Handle method directly before persisting the log to MongoDB.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	if err := h.handler.Handle(ctx, r); err != nil {
		return err
	}
	doc := bson.D{}
	if !r.Time.IsZero() {
		doc = append(doc, bson.E{Key: slog.TimeKey, Value: r.Time})
	}
	doc = append(doc, bson.E{Key: slog.LevelKey, Value: r.Level.String()})
	doc = append(doc, bson.E{Key: slog.MessageKey, Value: r.Message})
	r.Attrs(func(attr slog.Attr) bool {
		doc = appendAttrs(doc, attr)
		return true
	})
	if _, err := h.collection.InsertOne(context.WithoutCancel(ctx), doc); err != nil {
		return err
	}
	return nil
}

// WithAttrs implements the [slog.Handler] WithAttrs interface method.
// It calls the wrapped handler's WithAttrs method directly.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		handler:    h.handler.WithAttrs(attrs),
		collection: h.collection,
	}
}

// WithGroup implements the [slog.Handler] WithGroup interface method.
// It calls the wrapped handler's WithGroup method directly.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		handler:    h.handler.WithGroup(name),
		collection: h.collection,
	}
}

func appendAttrs(doc bson.D, attrs ...slog.Attr) bson.D {
	for _, attr := range attrs {
		if !attr.Equal(slog.Attr{}) {
			switch attr.Value.Kind() {
			case slog.KindAny:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Any()})
			case slog.KindBool:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Bool()})
			case slog.KindDuration:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Duration()})
			case slog.KindFloat64:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Float64()})
			case slog.KindInt64:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Int64()})
			case slog.KindString:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.String()})
			case slog.KindTime:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Time()})
			case slog.KindUint64:
				doc = append(doc, bson.E{Key: attr.Key, Value: attr.Value.Uint64()})
			case slog.KindGroup:
				groupAttrs := attr.Value.Group()
				if len(groupAttrs) > 0 {
					if attr.Key != "" {
						groupDoc := bson.D{}
						for _, groupAttr := range groupAttrs {
							groupDoc = appendAttrs(groupDoc, groupAttr)
						}
						doc = append(doc, bson.E{Key: attr.Key, Value: groupDoc})
					} else {
						doc = appendAttrs(doc, groupAttrs...)
					}
				}
			case slog.KindLogValuer:
				doc = appendAttrs(doc, slog.Attr{Key: attr.Key, Value: attr.Value.Resolve()})
			default:
				panic(fmt.Sprintf("bad kind: %v", attr.Value.Kind()))
			}
		}
	}
	return doc
}
