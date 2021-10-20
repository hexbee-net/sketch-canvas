package datastore

import (
	"context"

	"github.com/hexbee-net/sketch-canvas/pkg/canvas"
)

type StoreError string

func (s StoreError) Error() string {
	return string(s)
}

const NotFound = StoreError("not found")

//go:generate mockery --name DataStore
type DataStore interface {
	GetSize(ctx context.Context) (int64, error)
	GetDocList(cursor uint64, count int64, ctx context.Context) ([]string, uint64, error)
	SetDocument(key string, doc *canvas.Canvas, ctx context.Context) error
	GetDocument(key string, ctx context.Context) (*canvas.Canvas, error)
	DeleteDocument(key string, ctx context.Context) error
}
