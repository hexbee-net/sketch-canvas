package datastore

import "context"

type contextKey int

const (
	ContextKey contextKey = iota
)

//go:generate mockery --name DataStore
type DataStore interface {
	SetDocument(key string, value interface{}, ctx context.Context) error
}
