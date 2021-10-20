package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/xerrors"

	"github.com/hexbee-net/sketch-canvas/pkg/canvas"
	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
	datastoreMocks "github.com/hexbee-net/sketch-canvas/pkg/datastore/mocks"
	"github.com/hexbee-net/sketch-canvas/pkg/keygen"
	keygenMocks "github.com/hexbee-net/sketch-canvas/pkg/keygen/mocks"
)

func testServer(t *testing.T, keyGen keygen.KeyGen) Server {
	t.Helper()

	return Server{
		port:         0,
		srv:          &http.Server{},
		router:       mux.NewRouter(),
		storeOptions: &datastore.RedisOptions{},
		keygen:       keyGen,
	}
}

func jsonMarshal(t *testing.T, v canvas.Canvas) []byte {
	t.Helper()

	ret, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal test data")
	}
	return ret
}

func MiddlewareMockDatastore(t *testing.T) (mux.MiddlewareFunc, *datastoreMocks.DataStore) {
	t.Helper()

	store := datastoreMocks.DataStore{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), DatastoreContextKey, &store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, &store
}

func TestServer_createDocument(t *testing.T) {
	type storeCommand struct {
		key   string
		value interface{}
		ret   error
	}
	type args struct {
		body []byte
		cmd  storeCommand
	}
	type response struct {
		code int
		body string
	}
	tests := []struct {
		name      string
		args      args
		response  response
		checkBody bool
	}{
		{
			name: "ok",
			args: args{
				body: jsonMarshal(t, canvas.Canvas{}),
				cmd:  storeCommand{key: mock.Anything, value: mock.Anything, ret: nil},
			},
			response: response{
				code: http.StatusCreated,
				body: "/v1/123",
			},
			checkBody: true,
		},
		{
			name: "invalid parameters",
			args: args{
				body: []byte("invalid"),
				cmd:  storeCommand{key: mock.Anything, value: mock.Anything, ret: nil},
			},
			response: response{
				code: http.StatusBadRequest,
			},
			checkBody: false,
		},
		{
			name: "store error",
			args: args{
				body: jsonMarshal(t, canvas.Canvas{}),
				cmd:  storeCommand{key: mock.Anything, value: mock.Anything, ret: xerrors.New("FAILED")},
			},
			response: response{
				code: http.StatusInternalServerError,
			},
			checkBody: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyGenMock := keygenMocks.KeyGen{}
			server := testServer(t, &keyGenMock)

			mw, storeMock := MiddlewareMockDatastore(t)
			server.setupRoutes(server.router, mw)

			keyGenMock.On("Generate").Return("123")
			storeMock.On("SetDocument", tt.args.cmd.key, tt.args.cmd.value, mock.Anything).Return(tt.args.cmd.ret)

			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, httptest.NewRequest("POST", "/v1/", strings.NewReader(string(tt.args.body))))

			assert.Equal(t, tt.response.code, w.Code)
			if tt.checkBody {
				assert.Equal(t, tt.response.body, w.Body.String())
			}
		})
	}
}
