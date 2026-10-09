package rpc

import (
	"errors"
	"testing"
	"time"

	"github.com/jonog/redalert/core"
	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/events"
	pb "github.com/jonog/redalert/servicepb"
	"github.com/jonog/redalert/storage"
	"github.com/jonog/redalert/utils"
	"golang.org/x/net/context"
)

type eventStorageStub struct {
	events []*events.Event
	err    error
}

func (*eventStorageStub) Store(*events.Event) error             { return nil }
func (*eventStorageStub) Last() (*events.Event, error)          { return nil, nil }
func (s *eventStorageStub) GetRecent() ([]*events.Event, error) { return s.events, s.err }

func TestEventListReturnsRecentEventsForRequestedCheck(t *testing.T) {
	service := core.NewService()
	when := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	value := 3.5
	newest := &events.Event{Time: utils.RFCTime{Time: when}, Tags: map[string]string{"greenalert": ""}, Messages: []string{"recovered"}, Data: data.CheckResponse{Metrics: data.Metrics{"load": &value, "unknown": nil}, Metadata: data.Metadata{"region": "west"}}}
	older := &events.Event{Time: utils.RFCTime{Time: when.Add(-time.Minute)}, Tags: map[string]string{"redalert": ""}, Messages: []string{"failed"}}
	success := &events.Event{Time: utils.RFCTime{Time: when.Add(-2 * time.Minute)}, Tags: map[string]string{}, Messages: []string{"healthy"}}
	otherStore := storage.NewMemoryList(10)
	if err := otherStore.Store(&events.Event{Messages: []string{"must not leak"}}); err != nil {
		t.Fatal(err)
	}
	otherCheck := &core.Check{Data: pb.Check{ID: "other"}, Store: otherStore}
	if err := service.RegisterCheck(&core.Check{Data: pb.Check{ID: "target"}, Store: &eventStorageStub{events: []*events.Event{newest, older, success}}}, nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := service.RegisterCheck(otherCheck, nil, 1); err != nil {
		t.Fatal(err)
	}
	response, err := (&server{service: service}).EventList(context.Background(), &pb.EventListRequest{ID: "target"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Events) != 3 {
		t.Fatalf("got %d events", len(response.Events))
	}
	if response.Events[0].Time != when.Format(time.RFC3339Nano) || response.Events[1].Time != when.Add(-time.Minute).Format(time.RFC3339Nano) || response.Events[2].Time != when.Add(-2*time.Minute).Format(time.RFC3339Nano) {
		t.Fatalf("event order/times: %#v", response.Events)
	}
	first := response.Events[0]
	if first.Tags["greenalert"] != "" || len(first.Messages) != 1 || first.Messages[0] != "recovered" {
		t.Fatalf("untagged recovery not preserved: %#v", first)
	}
	if len(first.Metrics) != 2 {
		t.Fatalf("metrics: %#v", first.Metrics)
	}
	metricValues := map[string]*pb.Metric{}
	for _, metric := range first.Metrics {
		metricValues[metric.Name] = metric
	}
	if metricValues["load"] == nil || !metricValues["load"].Present || metricValues["load"].Value != value || metricValues["unknown"] == nil || metricValues["unknown"].Present {
		t.Fatalf("nullable metrics: %#v", metricValues)
	}
	if first.Metadata["region"] != "west" {
		t.Fatalf("metadata: %#v", first.Metadata)
	}
	if len(response.Events[2].Tags) != 0 || response.Events[2].Messages[0] != "healthy" {
		t.Fatalf("untagged successful event missing: %#v", response.Events[2])
	}
}

func TestEventListEmptyUnknownAndStorageErrors(t *testing.T) {
	service := core.NewService()
	empty := &core.Check{Data: pb.Check{ID: "empty"}, Store: storage.NewMemoryList(10)}
	if err := service.RegisterCheck(empty, nil, 0); err != nil {
		t.Fatal(err)
	}
	response, err := (&server{service: service}).EventList(context.Background(), &pb.EventListRequest{ID: "empty"})
	if err != nil || response == nil || len(response.Events) != 0 {
		t.Fatalf("empty result = %#v, %v", response, err)
	}
	if _, err := (&server{service: service}).EventList(context.Background(), &pb.EventListRequest{ID: "missing"}); err == nil {
		t.Fatal("unknown check should fail")
	}
	storageErr := errors.New("storage unavailable")
	failing := &core.Check{Data: pb.Check{ID: "failing"}, Store: &eventStorageStub{err: storageErr}}
	if err := service.RegisterCheck(failing, nil, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := (&server{service: service}).EventList(context.Background(), &pb.EventListRequest{ID: "failing"}); !errors.Is(err, storageErr) {
		t.Fatalf("storage error = %v", err)
	}
}
