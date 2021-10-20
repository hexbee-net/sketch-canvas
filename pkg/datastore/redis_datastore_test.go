package datastore

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
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
				t.Errorf("SetDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.expected, size)
		})
	}
}

func TestRedisDataStore_SetDocument(t *testing.T) {
	type args struct {
		key   string
		value interface{}
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
				value: struct{ value int }{value: 1},
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
				value: struct{ value int }{value: 1},
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
				value: struct{ value int }{value: 1},
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
				t.Errorf("SetDocument() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.expected.keys, keys)
			assert.Equal(t, tt.expected.cursor, cursor)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
