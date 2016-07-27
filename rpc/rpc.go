package rpc

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

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
