package datastore

import "context"

//go:generate mockery --name DataStore
type DataStore interface {
	SetDocument(key string, value interface{}, ctx context.Context) error
}
