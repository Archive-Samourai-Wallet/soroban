package soroban

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestRedis_Put(t *testing.T) {
	t.Parallel()
	r := NewRedis(OptionRedis{"127.0.0.1", 6379})

	key := randSeq(12)
	value := randSeq(12)

	type fields struct {
		r *Redis
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
		{"KeyErr", fields{r}, args{"", "value", time.Second}, true},
		{"ValueErr", fields{r}, args{"key", "", time.Second}, true},
		{"TTLOneErr", fields{r}, args{"key", "value", time.Second - time.Nanosecond}, true},
		{"TTLZeroErr", fields{r}, args{"key", "value", 0}, true},
		{"TTLNegativeErr", fields{r}, args{"key", "value", -time.Second}, true},
		{"Add", fields{r}, args{key, value, 1 * time.Second}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.r

			if err := r.Put(tt.args.key, tt.args.value, tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("Redis.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := r.Put(tt.args.key, tt.args.value, tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("Redis.Put() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			values, _ := r.Get(tt.args.key)
			if len(values) == 0 {
				t.Errorf("Unable to retrieve value")
			}
		})
	}
}

func TestRedis_Exists(t *testing.T) {
	t.Parallel()
	r := NewRedis(OptionRedis{"127.0.0.1", 6379})

	key := randSeq(12)
	value := randSeq(12)

	_ = r.Put(key, value, 1*time.Second)

	type fields struct {
		r *Redis
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"KeyErr", fields{r}, args{""}, false},
		{"NotExists", fields{r}, args{randSeq(12)}, false},
		{"Exists", fields{r}, args{key}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.r

			got := r.Exists(tt.args.key)
			if got != tt.want {
				t.Errorf("Redis.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedis_Get(t *testing.T) {
	t.Parallel()
	r := NewRedis(OptionRedis{"127.0.0.1", 6379})

	key := randSeq(12)
	value := randSeq(12)

	_ = r.Put(key, value, 1*time.Second)

	type fields struct {
		r *Redis
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
		{"KeyErr", fields{r}, args{""}, nil, true},
		{"Unknown", fields{r}, args{randSeq(12)}, nil, false},
		{"Get", fields{r}, args{key}, []string{value}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.r

			got, err := r.Get(tt.args.key)
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
			got, _ = r.Get(tt.args.key)
			if len(got) != 0 {
				t.Errorf("Getting an exipred key")
			}
		})
	}
}

func TestRedis_Del(t *testing.T) {
	t.Parallel()
	r := NewRedis(OptionRedis{"127.0.0.1", 6379})

	key := randSeq(12)
	value := randSeq(12)

	_ = r.Put(key, value, 1*time.Second)

	type fields struct {
		r *Redis
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"KeyErr", fields{r}, args{""}, true},
		{"Del", fields{r}, args{key}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.r

			if !tt.wantErr {
				got, _ := r.Get(tt.args.key)
				if len(got) == 0 {
					t.Errorf("Key not found")
				}
			}

			if err := r.Del(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Redis.Del() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got, _ := r.Get(tt.args.key)
			if len(got) != 0 {
				t.Errorf("Getting a deleted key")
			}
		})
	}
}
