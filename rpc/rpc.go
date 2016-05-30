package rpc

import (
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/jonog/redalert/core"
	pb "github.com/jonog/redalert/service"
	"google.golang.org/grpc"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	service *core.Service
}

func (s *server) ListChecks(ctx context.Context, in *pb.ListChecksRequest) (*pb.ListChecksResponse, error) {

	checks := s.service.Checks()
	rpcChecks := make([]*pb.Check, len(checks))

	for idx, check := range checks {

		var status pb.Check_Status
		event, err := check.Store.Last()
		if event != nil && err == nil {

			if event.IsRedAlert() {
				status = pb.Check_FAILING

			} else if event.IsGreenAlert() {
				status = pb.Check_RECOVERED

			} else {
				status = pb.Check_NORMAL
			}
		}

		rpcChecks[idx] = &pb.Check{
			ID:     check.ID,
			Name:   check.Name,
			Status: status,
		}
	}

	return &pb.ListChecksResponse{Members: rpcChecks}, nil
}

func Run(service *core.Service) {

	if os.Getenv("GRPC_TRACING_ENABLED") != "" {
		// Access trace via localhost:8080/debug/requests
		grpc.EnableTracing = true
		go http.ListenAndServe(":8080", nil)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRedalertServiceServer(s, &server{service})
	s.Serve(lis)
}
