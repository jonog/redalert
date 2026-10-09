package rpc

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/jonog/redalert/core"
	pb "github.com/jonog/redalert/servicepb"
	"google.golang.org/grpc"
)

type server struct {
	service *core.Service
}

func (s *server) CheckList(ctx context.Context, in *pb.CheckListRequest) (*pb.CheckListResponse, error) {
	checks := s.service.Checks()
	rpcChecks := make([]*pb.Check, len(checks))
	for idx, check := range checks {
		rpcChecks[idx] = &check.Data
	}
	return &pb.CheckListResponse{Members: rpcChecks}, nil
}

func (s *server) EventList(ctx context.Context, in *pb.EventListRequest) (*pb.EventListResponse, error) {
	check, err := s.service.CheckByID(in.ID)
	if err != nil {
		return nil, err
	}
	events, err := check.Store.GetRecent()
	if err != nil {
		return nil, err
	}
	response := &pb.EventListResponse{Events: make([]*pb.Event, 0, len(events))}
	for _, event := range events {
		item := &pb.Event{Time: event.Time.Time.Format(time.RFC3339Nano), Tags: event.Tags, Messages: event.Messages, Metadata: event.Data.Metadata}
		for name, value := range event.Data.Metrics {
			metric := &pb.Metric{Name: name}
			if value != nil {
				metric.Value = *value
				metric.Present = true
			}
			item.Metrics = append(item.Metrics, metric)
		}
		response.Events = append(response.Events, item)
	}
	return response, nil
}

func (s *server) CheckEnable(ctx context.Context, in *pb.CheckEnableRequest) (*pb.CheckEnableResponse, error) {

	check, err := s.service.CheckByID(in.ID)
	if err != nil {
		return nil, err
	}

	if check.Data.Enabled {
		return nil, errors.New("Check is already enabled")
	}

	go check.Start()

	return &pb.CheckEnableResponse{}, nil
}

func (s *server) CheckDisable(ctx context.Context, in *pb.CheckDisableRequest) (*pb.CheckDisableResponse, error) {

	check, err := s.service.CheckByID(in.ID)
	if err != nil {
		return nil, err
	}

	if !check.Data.Enabled {
		return nil, errors.New("Check is already disabled")
	}

	check.Stop()

	return &pb.CheckDisableResponse{}, nil
}

func Run(service *core.Service, port int) {

	if os.Getenv("GRPC_TRACING_ENABLED") != "" {
		// Access trace via localhost:8080/debug/requests
		grpc.EnableTracing = true
		go http.ListenAndServe(":8080", nil)
	}

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRedalertServiceServer(s, &server{service})
	s.Serve(lis)
}
