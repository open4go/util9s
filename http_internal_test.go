package util9s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

func TestInternalTokenOnlyForRPCPaths(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.URL.Path+"="+r.Header.Get(InternalTokenHeader))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	prev := viper.Get("rpc.auth_token")
	viper.Set("rpc.auth_token", "s3cret")
	defer viper.Set("rpc.auth_token", prev)

	ctx := context.Background()
	var rsp map[string]any
	if err := Get(ctx, srv.URL+"/rpc/store/1", &rsp); err != nil {
		t.Fatal(err)
	}
	if _, err := Post(ctx, srv.URL+"/v1/hlj/device/rpc/print", map[string]string{"a": "b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Post(ctx, srv.URL+"/external/api", nil); err != nil {
		t.Fatal(err)
	}
	if err := Get(ctx, srv.URL+"/x?next=/rpc/", &rsp); err != nil {
		t.Fatal(err)
	}
	want := []string{"/rpc/store/1=s3cret", "/v1/hlj/device/rpc/print=s3cret", "/external/api=", "/x="}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("req %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestInternalTokenLegacyKey(t *testing.T) {
	prevA, prevB := viper.Get("rpc.auth_token"), viper.Get("rpc.token")
	viper.Set("rpc.auth_token", "")
	viper.Set("rpc.token", "old")
	defer func() { viper.Set("rpc.auth_token", prevA); viper.Set("rpc.token", prevB) }()
	if internalToken() != "old" {
		t.Fatal("legacy key")
	}
}
