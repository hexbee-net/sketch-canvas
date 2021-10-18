package datastore

import (
	"context"
	"github.com/go-redis/redis/v8"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		options *redis.Options
		ctx     context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *DataStore
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
				options: &redis.Options{
					Addr: "localhost:6379",
				},
				ctx: nil,
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "Server down",
			args: args{
				options: &redis.Options{
					Addr: "invalid",
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
