package rpc

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonog/redalert/config"
	"github.com/jonog/redalert/core"
	pb "github.com/jonog/redalert/servicepb"
	"golang.org/x/net/context"
)

func TestCheckAddRejectsMalformedInputAndDestinationWithoutWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	initial := []byte(`{"checks":[],"notifications":[],"preferences":{}}`)
	if err := ioutil.WriteFile(path, initial, 0600); err != nil {
		t.Fatal(err)
	}
	store, err := config.NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	initial, err = ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := &server{service: core.NewService(), store: store, filename: path}
	if _, err = srv.CheckAdd(context.Background(), &pb.CheckAddRequest{Destination: path, Json: "{"}); err == nil {
		t.Fatal("malformed JSON accepted")
	}
	if _, err = srv.CheckAdd(context.Background(), &pb.CheckAddRequest{Destination: path + ".other", Json: "[]"}); err == nil {
		t.Fatal("destination mismatch accepted")
	}
	got, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(initial) {
		t.Fatalf("configuration changed after rejected additions: %s", got)
	}
}

func TestCheckAddPersistsRegistersAndStartsCheck(t *testing.T) {
	requests := make(chan struct{}, 1)
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case requests <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer fixture.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := ioutil.WriteFile(path, []byte(`{"checks":[],"notifications":[],"preferences":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := config.NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	service := core.NewService()
	srv := &server{service: service, store: store, filename: path}
	input := `[{"name":"fixture","type":"web-ping","config":{"address":"` + fixture.URL + `"},"backoff":{"type":"constant","interval":1},"assertions":[]}]`
	response, err := srv.CheckAdd(context.Background(), &pb.CheckAddRequest{Destination: path, Json: input})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Ids) != 1 || response.Ids[0] == "" {
		t.Fatalf("added IDs = %#v", response.Ids)
	}
	if _, err = service.CheckByID(response.Ids[0]); err != nil {
		t.Fatalf("check not registered: %v", err)
	}
	select {
	case <-requests:
	case <-time.After(8 * time.Second):
		t.Fatal("check did not issue a request to the local fixture")
	}
	check, _ := service.CheckByID(response.Ids[0])
	check.Stop()
	reloaded, err := config.NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	checks, err := reloaded.Checks()
	if err != nil || len(checks) != 1 || checks[0].ID != response.Ids[0] {
		t.Fatalf("persisted checks = %#v, err=%v", checks, err)
	}
}
