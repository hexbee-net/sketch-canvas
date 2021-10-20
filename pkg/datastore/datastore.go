package datastore

import (
	"context"

	"github.com/hexbee-net/sketch-canvas/pkg/canvas"
)

//go:generate mockery --name DataStore
type DataStore interface {
	GetSize(ctx context.Context) (int64, error)
	GetDocList(cursor uint64, count int64, ctx context.Context) ([]string, uint64, error)
	SetDocument(key string, doc *canvas.Canvas, ctx context.Context) error
	GetDocument(key string, ctx context.Context) (*canvas.Canvas, error)
}
