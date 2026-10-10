package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

func TestURLStoreYAMLUsesPathBeforeQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("checks: []\nnotifications: []\npreferences: {}\n"))
	}))
	defer server.Close()
	store, err := NewURLStore(server.URL + "/config.yaml?format=json#part")
	if err != nil {
		t.Fatal(err)
	}
	if checks, _ := store.Checks(); len(checks) != 0 {
		t.Fatalf("checks = %#v", checks)
	}
}

func TestS3StoreYAML(t *testing.T) {
	old := fetchS3File
	fetchS3File = func(_ *session.Session, bucket, key string) ([]byte, error) {
		if bucket != "bucket" || key != "/config.yml" {
			t.Fatalf("bucket/key = %q %q", bucket, key)
		}
		return []byte("checks: []\nnotifications: []\npreferences: {}\n"), nil
	}
	defer func() { fetchS3File = old }()
	store, err := NewS3Store("s3://bucket/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if checks, _ := store.Checks(); len(checks) != 0 {
		t.Fatalf("checks = %#v", checks)
	}
}
