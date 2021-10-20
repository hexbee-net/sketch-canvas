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
	keygenMocks "github.com/hexbee-net/sketch-canvas/pkg/keygen/mocks"
)

func testServer(t *testing.T) testSrv {
	t.Helper()

	keyGen := &keygenMocks.KeyGen{}
	mw, storeMock := MiddlewareMockDatastore(t)

	server := &Server{
		port:         0,
		srv:          &http.Server{},
		router:       mux.NewRouter(),
		storeOptions: &datastore.RedisOptions{},
		keygen:       keyGen,
	}

	server.router.Use(MiddlewareRequestID)
	server.setupRoutes(server.router, mw)

	return testSrv{
		keyGenMock: keyGen,
		storeMock:  storeMock,
		server:     server,
	}
}

func toJson(t *testing.T, v canvas.Canvas) []byte {
	t.Helper()

	ret, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal test data")
	}
	return ret
}

func MiddlewareMockDatastore(t *testing.T) (mux.MiddlewareFunc, *datastoreMocks.DataStore) {
	t.Helper()

	store := &datastoreMocks.DataStore{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), DatastoreContextKey, store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, store
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
				body: toJson(t, canvas.Canvas{}),
				cmd:  storeCommand{key: mock.Anything, value: mock.Anything, ret: nil},
			},
			response: response{
				code: http.StatusCreated,
				body: "/v1/docs/123",
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
				body: toJson(t, canvas.Canvas{}),
				cmd:  storeCommand{key: mock.Anything, value: mock.Anything, ret: xerrors.New("FAILED")},
			},
			response: response{
				code: http.StatusInternalServerError,
			},
			checkBody: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSrv := testServer(t)

			testSrv.keyGenMock.On("Generate").Return("123")
			testSrv.storeMock.On("SetDocument", tt.args.cmd.key, tt.args.cmd.value, mock.Anything).Return(tt.args.cmd.ret)
			w := httptest.NewRecorder()

			testSrv.server.router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/docs/", strings.NewReader(string(tt.args.body))))

			assert.Equal(t, tt.response.code, w.Code)
			if tt.checkBody {
				assert.Equal(t, tt.response.body, w.Body.String())
			}
		})
	}
}

func TestServer_getDocumentList(t *testing.T) {
	type storeGetList struct {
		cursor    uint64
		count     int64
		keys      []string
		newCursor uint64
		err       error
	}
	type storeGetSize struct {
		size int64
		err  error
	}
	type args struct {
		query string
	}
	type response struct {
		dbSize uint64
		code   int
		body   string
	}
	tests := []struct {
		name         string
		args         args
		storeGetList storeGetList
		storeGetSize storeGetSize
		response     response
		checkBody    bool
	}{
		{
			name: "ok",
			args: args{
				query: "?q=0&limit=5",
			},
			storeGetList: storeGetList{
				cursor:    0,
				count:     5,
				keys:      []string{"123", "456"},
				newCursor: 7,
				err:       nil,
			},
			storeGetSize: storeGetSize{
				size: 2,
				err:  nil,
			},
			response: response{
				code: http.StatusOK,
				body: `{"next":"/v1/docs/?limit=5&q=7","count":2,"total":2,"docs":{"123":"/v1/docs/123","456":"/v1/docs/456"}}`,
			},
			checkBody: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSrv := testServer(t)

			testSrv.storeMock.On("GetDocList", tt.storeGetList.cursor, tt.storeGetList.count, mock.Anything).Return(tt.storeGetList.keys, tt.storeGetList.newCursor, tt.storeGetList.err)
			testSrv.storeMock.On("GetSize", mock.Anything).Return(tt.storeGetSize.size, tt.storeGetSize.err)
			w := httptest.NewRecorder()

			testSrv.server.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/docs/"+tt.args.query, strings.NewReader("")))

			assert.Equal(t, tt.response.code, w.Code)
			if tt.checkBody {
				assert.Equal(t, tt.response.body+"\n", w.Body.String())
			}
		})
	}
}

type testSrv struct {
	keyGenMock *keygenMocks.KeyGen
	storeMock  *datastoreMocks.DataStore
	server     *Server
}
