package server

import (
	"context"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/mux"

	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
)

const epoch int64 = 1288834974657

func MiddlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curTime := time.Now()
		id := time.Since(curTime.Add(time.Unix(epoch/1000, (epoch%1000)*1000000).Sub(curTime))).Nanoseconds() / 1000000 //nolint:gomnd

		log.Debugf("new request id: %v", id)

		ctx := context.WithValue(r.Context(), RequestIDContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MiddlewareRedisDatastore(storeOptions *datastore.RedisOptions) mux.MiddlewareFunc {
	if storeOptions == nil {
		log.Fatal("store options cannot be nil in middleware")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			store, err := datastore.New(storeOptions, r.Context())
			if err != nil {
				log.WithError(err).Error("datastore is unreachable")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			ctx := context.WithValue(r.Context(), DatastoreContextKey, store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
