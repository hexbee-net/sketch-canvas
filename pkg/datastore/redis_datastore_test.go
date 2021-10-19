package datastore

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
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
