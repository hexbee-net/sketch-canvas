package datastore

import (
	"context"
	"net/http"

	"github.com/apex/log"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

type RedisOptions struct {
	redis.Options
}

type RedisDataStore struct {
	rdb *redis.Client
}

func MiddlewareRedisDatastore(storeOptions *RedisOptions) mux.MiddlewareFunc {
	if storeOptions == nil {
		log.Fatal("store options cannot be nil in middleware")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			store, err := New(storeOptions, r.Context())
			if err != nil {
				log.WithError(err).Error("datastore is unreachable")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			ctx := context.WithValue(r.Context(), ContextKey, store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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

func (s *RedisDataStore) SetDocument(key string, value interface{}, ctx context.Context) error {
	if err := s.rdb.Set(ctx, key, value, 0).Err(); err != nil {
		return xerrors.Errorf("failed to set document in redis store: %w", err)
	}

	return nil
}
