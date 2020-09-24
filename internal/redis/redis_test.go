package redis

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestRedis_Create(t *testing.T) {
	t.Parallel()
	directory := New(soroban.ServerInfo{Hostname: "127.0.0.1", Port: 6379})

	key := randSeq(12)
	value := randSeq(12)

	type fields struct {
		directory soroban.Directory
	}
	type args struct {
		key   string
		value string
		ttl   time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"KeyErr", fields{directory}, args{"", "value", time.Second}, true},
		{"ValueErr", fields{directory}, args{"key", "", time.Second}, true},
		{"TTLOneErr", fields{directory}, args{"key", "value", time.Second - time.Nanosecond}, true},
		{"TTLZeroErr", fields{directory}, args{"key", "value", 0}, true},
		{"TTLNegativeErr", fields{directory}, args{"key", "value", -time.Second}, true},
		{"Add", fields{directory}, args{key, value, 1 * time.Second}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			directory := tt.fields.directory

			if err := directory.Add(tt.args.key, tt.args.value, tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("Redis.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := directory.Add(tt.args.key, tt.args.value, tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("Redis.Put() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			values, _ := directory.List(tt.args.key)
			if len(values) == 0 {
				t.Errorf("Unable to retrieve value")
			}
		})
	}
}

func TestRedis_List(t *testing.T) {
	t.Parallel()
	directory := New(soroban.ServerInfo{Hostname: "127.0.0.1", Port: 6379})

	key := randSeq(12)

	// store multiple values and check order is kept
	var values []string
	for i := 0; i < 10; i++ {
		value := randSeq(12)
		_ = directory.Add(key, value, 1*time.Second)
		values = append(values, value)
	}

	type fields struct {
		directory soroban.Directory
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{"KeyErr", fields{directory}, args{""}, nil, true},
		{"Unknown", fields{directory}, args{randSeq(12)}, nil, false},
		{"Get", fields{directory}, args{key}, values, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			directory := tt.fields.directory

			got, err := directory.List(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redis.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Redis.Get() = %v, want %v", got, tt.want)
			}

			if len(got) == 0 {
				return
			}

			time.Sleep(1200 * time.Millisecond)
			got, _ = directory.List(tt.args.key)
			if len(got) != 0 {
				t.Errorf("Getting an exipred key")
			}
		})
	}
}

func TestRedis_Delete(t *testing.T) {
	t.Parallel()
	directory := New(soroban.ServerInfo{Hostname: "127.0.0.1", Port: 6379})

	key := randSeq(12)
	value := randSeq(12)

	_ = directory.Add(key, value, 1*time.Second)

	type fields struct {
		directory soroban.Directory
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"KeyErr", fields{directory}, args{"", ""}, true},
		{"Del", fields{directory}, args{key, value}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			directory := tt.fields.directory

			if !tt.wantErr {
				got, _ := directory.List(tt.args.key)
				if len(got) == 0 {
					t.Errorf("Key not found")
				}
			}

			if err := directory.Remove(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Redis.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got, _ := directory.List(tt.args.key)
			if len(got) != 0 {
				t.Errorf("Getting a deleted key")
			}
		})
	}
}
