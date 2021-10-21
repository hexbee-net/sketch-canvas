package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
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

type testSrv struct {
	keyGenMock *keygenMocks.KeyGen
	storeMock  *datastoreMocks.DataStore
	server     *Server
}

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

func TestServer_getDocument(t *testing.T) {
	type storeGetDocument struct {
		docID string
		doc   *canvas.Canvas
		err   error
	}
	type response struct {
		code int
		body string
	}
	tests := []struct {
		name             string
		storeGetDocument storeGetDocument
		response         response
		checkBody        bool
	}{
		{
			name: "ok",
			storeGetDocument: storeGetDocument{
				docID: "123",
				doc:   &canvas.Canvas{Name: "doc1", Width: 80, Height: 50},
				err:   nil,
			},
			response: response{
				code: http.StatusOK,
				body: `{"operations":{"add-flood-fill":"/v1/docs/123/fill","add-rect":"/v1/docs/123/rect","delete-doc":"/v1/docs/123"},"Canvas":{"name":"doc1","width":80,"height":50}}`,
			},
			checkBody: true,
		},
		{
			name: "not found",
			storeGetDocument: storeGetDocument{
				docID: "123",
				doc:   nil,
				err:   datastore.NotFound,
			},
			response: response{
				code: http.StatusNotFound,
				body: ``,
			},
			checkBody: false,
		},
		{
			name: "store error",
			storeGetDocument: storeGetDocument{
				docID: "123",
				doc:   nil,
				err:   xerrors.New("FAILED"),
			},
			response: response{
				code: http.StatusInternalServerError,
				body: ``,
			},
			checkBody: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSrv := testServer(t)

			testSrv.storeMock.On("GetDocument", tt.storeGetDocument.docID, mock.Anything).Return(tt.storeGetDocument.doc, tt.storeGetDocument.err)
			w := httptest.NewRecorder()

			testSrv.server.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/docs/123", strings.NewReader("")))

			assert.Equal(t, tt.response.code, w.Code)
			if tt.checkBody {
				assert.Equal(t, tt.response.body+"\n", w.Body.String())
			}
		})
	}
}

func TestServer_deleteDocument(t *testing.T) {
	type storeDeleteDocument struct {
		docID string
		err   error
	}
	tests := []struct {
		name                string
		storeDeleteDocument storeDeleteDocument
		response            int
		checkBody           bool
	}{
		{
			name: "ok",
			storeDeleteDocument: storeDeleteDocument{
				docID: "123",
				err:   nil,
			},
			response:  http.StatusNoContent,
			checkBody: true,
		},
		{
			name: "not found",
			storeDeleteDocument: storeDeleteDocument{
				docID: "123",
				err:   datastore.NotFound,
			},
			response:  http.StatusNotFound,
			checkBody: false,
		},
		{
			name: "store error",
			storeDeleteDocument: storeDeleteDocument{
				docID: "123",
				err:   xerrors.New("FAILED"),
			},
			response:  http.StatusInternalServerError,
			checkBody: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSrv := testServer(t)

			testSrv.storeMock.On("DeleteDocument", tt.storeDeleteDocument.docID, mock.Anything).Return(tt.storeDeleteDocument.err)
			w := httptest.NewRecorder()

			testSrv.server.router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/v1/docs/123", strings.NewReader("")))

			assert.Equal(t, tt.response, w.Code)
		})
	}
}

func TestServer_Operations(t *testing.T) {
	type storeGetCommand struct {
		docID string
		doc   *canvas.Canvas
		err   error
	}
	type storeSetCommand struct {
		docID string
		doc   interface{}
		err   error
	}
	type args struct {
		operation string
		body      string
	}
	type response struct {
		code int
	}
	tests := []struct {
		name       string
		args       args
		getCommand storeGetCommand
		setCommand storeSetCommand
		response   response
		checkBody  bool
	}{
		{
			name: "rect ok",
			args: args{
				operation: "rect",
				body:      `{"rect":{"origin":{"x":2,"y":3},"width":4,"height":5},"fill":"X","outline":"@"}`,
			},
			getCommand: storeGetCommand{
				docID: mock.Anything,
				doc: &canvas.Canvas{
					Name:   "doc1",
					Width:  10,
					Height: 10,
					Data:   nil,
				},
				err: nil,
			},
			setCommand: storeSetCommand{
				docID: mock.Anything,
				doc:   mock.Anything,
				err:   nil,
			},
			response: response{
				code: http.StatusOK,
			},
			checkBody: true,
		},
		{
			name: "rect - missing parameters",
			args: args{
				operation: "rect",
				body:      `{"rect":{"origin":{"x":5,"y":5},"width":10,"height":4}}`,
			},
			getCommand: storeGetCommand{
				docID: mock.Anything,
				doc:   &canvas.Canvas{},
				err:   nil,
			},
			setCommand: storeSetCommand{
				docID: mock.Anything,
				doc:   mock.Anything,
				err:   nil,
			},
			response: response{
				code: http.StatusBadRequest,
			},
			checkBody: true,
		},
		{
			name: "fill ok",
			args: args{
				operation: "fill",
				body:      `{"origin":{"x":5,"y":5},"fill":"X"}`,
			},
			getCommand: storeGetCommand{
				docID: mock.Anything,
				doc: &canvas.Canvas{
					Name:   "doc1",
					Width:  10,
					Height: 10,
					Data:   nil,
				},
				err: nil,
			},
			setCommand: storeSetCommand{
				docID: mock.Anything,
				doc:   mock.Anything,
				err:   nil,
			},
			response: response{
				code: http.StatusOK,
			},
			checkBody: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSrv := testServer(t)

			testSrv.storeMock.On("GetDocument", tt.getCommand.docID, mock.Anything).Return(tt.getCommand.doc, tt.getCommand.err)
			testSrv.storeMock.On("SetDocument", tt.setCommand.docID, tt.setCommand.doc, mock.Anything).Return(tt.setCommand.err)
			w := httptest.NewRecorder()

			testSrv.server.router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path.Join("/v1/docs/123/", tt.args.operation), strings.NewReader(tt.args.body)))

			assert.Equal(t, tt.response.code, w.Code)
		})
	}
}
