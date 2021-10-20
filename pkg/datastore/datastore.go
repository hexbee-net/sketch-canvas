package datastore

import "context"

//go:generate mockery --name DataStore
type DataStore interface {
	GetSize(ctx context.Context) (int64, error)
	GetDocList(cursor uint64, count int64, ctx context.Context) ([]string, uint64, error)
	SetDocument(key string, value interface{}, ctx context.Context) error
}
