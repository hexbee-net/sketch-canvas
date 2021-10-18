package datastore

import (
	"context"
	"github.com/apex/log"
	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"
)

type Options struct {
	redis.Options
}

type DataStore struct {
	rdb *redis.Client
}

// New creates a new DataStore instance and check the connectivity to the Redis instance.
func New(options *Options, ctx context.Context) (*DataStore, error) {
	if options == nil {
		return nil, xerrors.New("missing Redis datastore configuration")
	}
	if ctx == nil {
		return nil, xerrors.New("missing execution context")
	}

	rdb := redis.NewClient(&options.Options)

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, xerrors.Errorf("failed to connect to Redis instance at %s: %w", options.Addr, err)
	}

	log.Debugf("found Redis instance at %s", options.Addr)

	store := DataStore{
		rdb: rdb,
	}

	return &store, nil
}
