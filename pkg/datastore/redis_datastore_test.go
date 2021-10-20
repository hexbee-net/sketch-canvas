package datastore

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"

	"github.com/hexbee-net/sketch-canvas/pkg/canvas"
)

func TestNew(t *testing.T) {
	type args struct {
		options *RedisOptions
		ctx     context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *RedisDataStore
		wantErr bool
	}{
		{
			name: "No datastore options",
			args: args{
				options: nil,
				ctx:     nil,
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "No execution context",
			args: args{
				options: &RedisOptions{
					redis.Options{
						Addr: "localhost:6379",
					},
				},
				ctx: nil,
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "Server down",
			args: args{
				options: &RedisOptions{
					redis.Options{
						Addr: "invalid",
					},
				},
				ctx: context.TODO(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.options, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisDataStore_GetSize(t *testing.T) {
	type command struct {
		val int64
		err error
	}
	tests := []struct {
		name     string
		cmd      command
		expected int64
		wantErr  bool
	}{
		{
			name: "redis OK",
			cmd: command{
				val: 12,
				err: nil,
			},
			expected: 12,
			wantErr:  false,
		},
		{
			name: "redis error",
			cmd: command{
				val: 0,
				err: xerrors.New("FAILED"),
			},
			expected: 0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			s := &RedisDataStore{
				rdb: db,
			}

			cmd := mock.ExpectDBSize()
			cmd.SetVal(tt.cmd.val)
			cmd.SetErr(tt.cmd.err)

			size, err := s.GetSize(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSize() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.expected, size)
		})
	}
}

func TestRedisDataStore_SetDocument(t *testing.T) {
	type args struct {
		key   string
		value *canvas.Canvas
	}
	type command struct {
		val string
		err error
	}
	tests := []struct {
		name    string
		args    args
		cmd     command
		wantErr bool
	}{
		{
			name: "redis OK",
			args: args{
				key:   "doc1",
				value: &canvas.Canvas{Name: "canvas1"},
			},
			cmd: command{
				val: "OK",
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "redis null",
			args: args{
				key:   "doc1",
				value: &canvas.Canvas{Name: "canvas1"},
			},
			cmd: command{
				val: "",
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "redis error",
			args: args{
				key:   "doc1",
				value: &canvas.Canvas{Name: "canvas1"},
			},
			cmd: command{
				val: "",
				err: xerrors.New("FAILED"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			s := &RedisDataStore{
				rdb: db,
			}

			cmd := mock.ExpectSet(tt.args.key, tt.args.value, 0)
			cmd.SetVal(tt.cmd.val)
			cmd.SetErr(tt.cmd.err)

			if err := s.SetDocument(tt.args.key, tt.args.value, context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("SetDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisDataStore_GetDocList(t *testing.T) {
	type args struct {
		cursor uint64
		count  int64
	}
	type command struct {
		page   []string
		cursor uint64
		err    error
	}
	type expected struct {
		keys   []string
		cursor uint64
	}
	tests := []struct {
		name     string
		args     args
		cmd      command
		expected expected
		wantErr  bool
	}{
		{
			name: "redis OK",
			args: args{
				cursor: 0,
				count:  10,
			},
			cmd: command{
				page:   []string{"123", "456"},
				cursor: 3,
				err:    nil,
			},
			expected: expected{
				keys:   []string{"123", "456"},
				cursor: 3,
			},
			wantErr: false,
		},
		{
			name: "redis error",
			args: args{
				cursor: 0,
				count:  10,
			},
			cmd: command{
				page:   []string{"123", "456"},
				cursor: 3,
				err:    xerrors.New("FAILED"),
			},
			expected: expected{
				keys:   nil,
				cursor: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			s := &RedisDataStore{
				rdb: db,
			}

			cmd := mock.ExpectScan(tt.args.cursor, "", tt.args.count)
			cmd.SetVal(tt.cmd.page, tt.cmd.cursor)
			cmd.SetErr(tt.cmd.err)

			keys, cursor, err := s.GetDocList(tt.args.cursor, tt.args.count, context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDocList() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.expected.keys, keys)
			assert.Equal(t, tt.expected.cursor, cursor)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisDataStore_GetDocument(t *testing.T) {
	type args struct {
		key string
	}
	type existCommand struct {
		value int64
		err   error
	}
	type getCommand struct {
		value string
		err   error
	}
	type expected struct {
		doc *canvas.Canvas
	}
	tests := []struct {
		name         string
		args         args
		existCommand existCommand
		getCommand   *getCommand
		expected     expected
		wantErr      bool
	}{
		{
			name: "ok",
			args: args{
				key: "123",
			},
			existCommand: existCommand{
				value: 1,
				err:   nil,
			},
			getCommand: &getCommand{
				value: `{"name":"doc1","width":80,"height":25,"data":"-#-"}`,
				err:   nil,
			},
			expected: expected{
				doc: &canvas.Canvas{
					Name:   "doc1",
					Width:  80,
					Height: 25,
					Data:   "-#-",
				},
			},
			wantErr: false,
		},
		{
			name: "exists error",
			args: args{
				"123",
			},
			existCommand: existCommand{
				value: 0,
				err:   xerrors.New("FAILED"),
			},
			getCommand: nil,
			expected: expected{
				doc: nil,
			},
			wantErr: true,
		},
		{
			name: "get error",
			args: args{
				"123",
			},
			existCommand: existCommand{
				value: 1,
				err:   nil,
			},
			getCommand: &getCommand{
				value: "",
				err:   xerrors.New("FAILED"),
			},
			expected: expected{
				doc: nil,
			},
			wantErr: true,
		},
		{
			name: "bad data",
			args: args{
				"123",
			},
			existCommand: existCommand{
				value: 1,
				err:   nil,
			},
			getCommand: &getCommand{
				value: "invalid",
				err:   nil,
			},
			expected: expected{
				doc: nil,
			},
			wantErr: true,
		},
		{
			name: "empty data",
			args: args{
				"123",
			},
			existCommand: existCommand{
				value: 1,
				err:   nil,
			},
			getCommand: &getCommand{
				value: "",
				err:   nil,
			},
			expected: expected{
				doc: nil,
			},
			wantErr: true,
		},
		{
			name: "not found",
			args: args{
				"123",
			},
			existCommand: existCommand{
				value: 0,
				err:   nil,
			},
			getCommand: nil,
			expected: expected{
				doc: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			s := &RedisDataStore{
				rdb: db,
			}

			expectExists := mock.ExpectExists(tt.args.key)
			expectExists.SetVal(tt.existCommand.value)
			expectExists.SetErr(tt.existCommand.err)

			if tt.getCommand != nil {
				expectGet := mock.ExpectGet(tt.args.key)
				expectGet.SetVal(tt.getCommand.value)
				expectGet.SetErr(tt.getCommand.err)
			}

			doc, err := s.GetDocument(tt.args.key, context.TODO())
			assert.Equal(t, tt.expected.doc, doc)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
