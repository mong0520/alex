package main

import (
	"testing"
)

func Test_randomString(t *testing.T) {
	got := randomString(5)
	t.Error(got)
}

func TestReplaceMapByEnvs(t *testing.T) {
	type args struct {
		envs    map[string]interface{}
		sources []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test1",
			args: args{
				envs: map[string]interface{}{"Authorization": "THIS_IS_A_PREDEFINED_AUTH_TOKEN"},
				sources: []map[string]interface{}{
					{"Authorization": "aBearer $Authorization", "Content-Type": "application/json"},
				},
			},
		},
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.args.sources)
			ReplaceMapByEnvs(tt.args.envs, idx, tt.args.sources...)
			t.Log(tt.args.sources)
		})
	}
}

func TestReplaceStringByEnvs(t *testing.T) {
	type args struct {
		envs   map[string]interface{}
		source *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				envs:   map[string]interface{}{"api": "crypto_bulkEncrypt"},
				source: NewString("{\"id\":!AUTONUM,\"jsonrpc\":\"2.0\",\"method\":\"$api\",\"params\":[[{\"workspace_id\":\"!UUID\",\"key_type\":\"HASH\",\"plaintext\":\"!B64RANDOM\"}]]}"),
			},
			wantErr: false,
		},
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Before = %s\n", *tt.args.source)
			if err := ReplaceStringByEnvs(tt.args.envs, idx, tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("ReplaceStringByEnvs() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("After = %s\n", *tt.args.source)
		})
	}
}

func NewString(s string) *string {
	return &s
}
