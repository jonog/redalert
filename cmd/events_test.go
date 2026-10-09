package cmd

import (
	"io"
	"net"
	"os"
	"strings"
	"testing"

	pb "github.com/jonog/redalert/servicepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

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
