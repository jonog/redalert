package cmd

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"testing"

	pb "github.com/jonog/redalert/servicepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

type eventTestServer struct {
	request  *pb.EventListRequest
	response *pb.EventListResponse
	err      error
}

func (*eventTestServer) CheckAdd(context.Context, *pb.CheckAddRequest) (*pb.CheckAddResponse, error) {
	return nil, nil
}

func (*eventTestServer) CheckList(context.Context, *pb.CheckListRequest) (*pb.CheckListResponse, error) {
	return nil, nil
}
func (*eventTestServer) CheckEnable(context.Context, *pb.CheckEnableRequest) (*pb.CheckEnableResponse, error) {
	return nil, nil
}
func (*eventTestServer) CheckDisable(context.Context, *pb.CheckDisableRequest) (*pb.CheckDisableResponse, error) {
	return nil, nil
}
func (s *eventTestServer) EventList(_ context.Context, in *pb.EventListRequest) (*pb.EventListResponse, error) {
	s.request = in
	return s.response, s.err
}

func startEventTestServer(t *testing.T, srv pb.RedalertServiceServer) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	pb.RegisterRedalertServiceServer(server, srv)
	go server.Serve(listener)
	t.Cleanup(func() { server.Stop(); listener.Close() })
	return listener.Addr().(*net.TCPAddr).Port
}

func TestEventsArgumentsAreValidatedBeforeRPC(t *testing.T) {
	for _, args := range [][]string{nil, {"one", "two"}} {
		if err := eventsCmd.RunE(eventsCmd, args); err == nil || !strings.Contains(err.Error(), "exactly one") {
			t.Fatalf("RunE(%v) error = %v, want argument error", args, err)
		}
	}
}

func TestEventsCommandRendersAllFieldsAndEmptyHistory(t *testing.T) {
	srv := &eventTestServer{response: &pb.EventListResponse{Events: []*pb.Event{{
		Time: "2025-01-02T03:04:05Z", Tags: map[string]string{"z": "last", "a": "first"},
		Messages: []string{"failed", "retrying"}, Metrics: []*pb.Metric{{Name: "nullable"}, {Name: "load", Value: 1.25, Present: true}},
		Metadata: map[string]string{"host": "node1"},
	}}}}
	rpcPort = startEventTestServer(t, srv)
	cmd := eventsCmd
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	if err := cmd.RunE(cmd, []string{"check-a"}); err != nil {
		t.Fatal(err)
	}
	w.Close()
	captured, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	got := string(captured)
	for _, want := range []string{"Time: 2025-01-02T03:04:05Z", "Tags: a=first, z=last", "Messages: failed; retrying", "Metrics: load=1.25, nullable=null", "Metadata: host=node1"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
	if srv.request == nil || srv.request.ID != "check-a" {
		t.Fatalf("RPC request = %#v", srv.request)
	}

	srv.response = &pb.EventListResponse{Events: []*pb.Event{}}
	// Use a fresh pipe to capture the empty-history response.
	r2, w2, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w2
	if err := cmd.RunE(cmd, []string{"check-a"}); err != nil {
		t.Fatal(err)
	}
	w2.Close()
	captured, err = io.ReadAll(r2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(captured), "No events found") {
		t.Fatalf("empty result = %q", string(captured))
	}
}

func TestEventsCommandReportsRPCAndUnavailableServerErrors(t *testing.T) {
	if err := eventsCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	defer eventsCmd.Flags().Set("json", "false")
	srv := &eventTestServer{err: context.DeadlineExceeded}
	rpcPort = startEventTestServer(t, srv)
	if err := eventsCmd.RunE(eventsCmd, []string{"missing"}); err == nil || !strings.Contains(err.Error(), "list events") {
		t.Fatalf("RPC error = %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	rpcPort = port
	if err := eventsCmd.RunE(eventsCmd, []string{"check-a"}); err == nil || !strings.Contains(err.Error(), "list events") {
		t.Fatalf("unavailable server error = %v", err)
	}
}

func TestEventsJSONOutputAndEmptyHistory(t *testing.T) {
	srv := &eventTestServer{response: &pb.EventListResponse{Events: []*pb.Event{{
		Time: "2025-01-02T03:04:05Z", Tags: map[string]string{"region": "west"},
		Messages: []string{"failed"}, Metrics: []*pb.Metric{{Name: "missing"}, {Name: "load", Value: 1.25, Present: true}, {Name: "zero", Value: 0, Present: true}},
		Metadata: map[string]string{"host": "node1"},
	}, {
		Time: "2025-01-03T04:05:06Z", Tags: map[string]string{"region": "east"},
		Messages: []string{"recovered"}, Metrics: []*pb.Metric{{Name: "load", Value: 2.5, Present: true}, {Name: "optional"}},
		Metadata: map[string]string{"host": "node2"},
	}}}}
	rpcPort = startEventTestServer(t, srv)
	if err := eventsCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	defer eventsCmd.Flags().Set("json", "false")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	eventsCmd.SetOutput(w)
	defer eventsCmd.SetOutput(nil)
	if err := eventsCmd.RunE(eventsCmd, []string{"check-a"}); err != nil {
		t.Fatal(err)
	}
	w.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	var got []*pb.Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("JSON output %q: %v", data, err)
	}
	var raw []struct {
		Metrics []map[string]json.RawMessage `json:"metrics"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("decode metric JSON fields: %v", err)
	}
	if _, ok := raw[0].Metrics[0]["present"]; !ok {
		t.Fatalf("false metric omitted present field: %s", data)
	}
	if string(raw[0].Metrics[0]["present"]) != "false" {
		t.Fatalf("false metric present field = %s, want false", raw[0].Metrics[0]["present"])
	}
	if string(raw[0].Metrics[2]["value"]) != "0" {
		t.Fatalf("zero metric value field = %s, want 0", raw[0].Metrics[2]["value"])
	}
	if len(got) != 2 || got[0].Time != "2025-01-02T03:04:05Z" || got[0].Tags["region"] != "west" ||
		len(got[0].Messages) != 1 || got[0].Messages[0] != "failed" || got[0].Metadata["host"] != "node1" ||
		len(got[0].Metrics) != 3 || got[0].Metrics[0].Present || !got[0].Metrics[1].Present || got[0].Metrics[1].Value != 1.25 || !got[0].Metrics[2].Present || got[0].Metrics[2].Value != 0 {
		t.Fatalf("decoded events = %#v", got)
	}
	if got[1].Time != "2025-01-03T04:05:06Z" || got[1].Tags["region"] != "east" ||
		len(got[1].Messages) != 1 || got[1].Messages[0] != "recovered" || got[1].Metadata["host"] != "node2" ||
		len(got[1].Metrics) != 2 || !got[1].Metrics[0].Present || got[1].Metrics[0].Value != 2.5 || got[1].Metrics[1].Present {
		t.Fatalf("second decoded event = %#v", got[1])
	}

	for _, events := range [][]*pb.Event{nil, {}} {
		srv.response = &pb.EventListResponse{Events: events}
		r2, w2, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		eventsCmd.SetOutput(w2)
		if err := eventsCmd.RunE(eventsCmd, []string{"check-a"}); err != nil {
			t.Fatal(err)
		}
		w2.Close()
		data, err := io.ReadAll(r2)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "[]\n" {
			t.Fatalf("empty JSON output = %q, want []\\n", data)
		}
	}
	eventsCmd.SetOutput(failingWriter{})
	if err := eventsCmd.RunE(eventsCmd, []string{"check-a"}); err == nil || !strings.Contains(err.Error(), "encode events as JSON") {
		t.Fatalf("output error = %v, want encoding/output error", err)
	}
}
