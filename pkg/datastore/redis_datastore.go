package datastore

import (
	"context"

	"github.com/apex/log"
	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"
)

type RedisOptions struct {
	redis.Options
}

type RedisDataStore struct {
	rdb *redis.Client
}

// New creates a new RedisDataStore instance and check the connectivity to the Redis instance.
func New(options *RedisOptions, ctx context.Context) (*RedisDataStore, error) {
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

	store := RedisDataStore{
		rdb: rdb,
	}

	return &store, nil
}

func (s *RedisDataStore) GetSize(ctx context.Context) (int64, error) {
	cmd := s.rdb.DBSize(ctx)
	if err := cmd.Err(); err != nil {
		return 0, xerrors.Errorf("failed to retrieve size of redis store: %w", err)
	}

	return cmd.Val(), nil
}

func (s *RedisDataStore) GetDocList(cursor uint64, count int64, ctx context.Context) ([]string, uint64, error) {
	cmd := s.rdb.Scan(ctx, cursor, "", count)
	if err := cmd.Err(); err != nil {
		return nil, 0, xerrors.Errorf("failed to retrieve documents from redis store: %w", err)
	}

	keys, cursor := cmd.Val()

	return keys, cursor, nil
}

func (s *RedisDataStore) SetDocument(key string, value interface{}, ctx context.Context) error {
	if err := s.rdb.Set(ctx, key, value, 0).Err(); err != nil {
		return xerrors.Errorf("failed to set document in redis store: %w", err)
	}

	return nil
}
