package rpc

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/config"
	"github.com/jonog/redalert/core"
	pb "github.com/jonog/redalert/servicepb"
	"github.com/jonog/redalert/storage"
	"google.golang.org/grpc"
)

type server struct {
	service  *core.Service
	store    config.Store
	filename string
	addMu    sync.Mutex
}

func (s *server) CheckAdd(ctx context.Context, in *pb.CheckAddRequest) (*pb.CheckAddResponse, error) {
	s.addMu.Lock()
	defer s.addMu.Unlock()
	var appendStore interface{ AppendChecks([]checks.Config) error }
	var ok bool
	appendStore, ok = s.store.(interface{ AppendChecks([]checks.Config) error })
	if !ok {
		return nil, errors.New("check-add requires file configuration")
	}
	dest, err := filepath.Abs(in.Destination)
	if err != nil {
		return nil, err
	}
	configured, err := filepath.Abs(s.filename)
	if err != nil {
		return nil, err
	}
	if dest != configured {
		return nil, fmt.Errorf("destination %q does not match server config file %q", in.Destination, s.filename)
	}
	var batch []checks.Config
	if err = json.Unmarshal([]byte(in.Json), &batch); err != nil {
		return nil, fmt.Errorf("invalid check JSON: %w", err)
	}
	if len(batch) == 0 {
		return nil, errors.New("check input must be a nonempty JSON array")
	}
	current, err := s.store.Checks()
	if err != nil {
		return nil, err
	}
	prefs, err := s.store.Preferences()
	if err != nil {
		return nil, err
	}
	known := make(map[string]bool)
	for _, c := range current {
		known[c.ID] = true
	}
	prepared := make([]*core.Check, 0, len(batch))
	notifications := make([][]string, 0, len(batch))
	ids := make([]string, 0, len(batch))
	for i := range batch {
		c := &batch[i]
		if c.Name == "" || c.Type == "" {
			return nil, fmt.Errorf("check %d requires name and type", i+1)
		}
		if c.Enabled != nil && !*c.Enabled {
			return nil, fmt.Errorf("check %q cannot be added disabled", c.Name)
		}
		if c.Backoff.Type != "" && c.Backoff.Type != "constant" && c.Backoff.Type != "linear" && c.Backoff.Type != "exponential" {
			return nil, fmt.Errorf("check %q has unsupported backoff type %q", c.Name, c.Backoff.Type)
		}
		if c.Backoff.Interval != nil && *c.Backoff.Interval <= 0 {
			return nil, fmt.Errorf("check %q backoff interval must be positive", c.Name)
		}
		if c.Backoff.Multiplier != nil && *c.Backoff.Multiplier <= 0 {
			return nil, fmt.Errorf("check %q backoff multiplier must be positive", c.Name)
		}
		if c.ID == "" {
			var id string
			for {
				b := make([]byte, 6)
				if _, err = rand.Read(b); err != nil {
					return nil, err
				}
				id = fmt.Sprintf("%x", b)
				_, runtimeExists := s.service.CheckByID(id)
				if !known[id] && runtimeExists != nil {
					break
				}
			}
			c.ID = id
		}
		if known[c.ID] {
			return nil, fmt.Errorf("duplicate check ID %q", c.ID)
		}
		if _, err := s.service.CheckByID(c.ID); err == nil {
			return nil, fmt.Errorf("check ID %q already exists at runtime", c.ID)
		}
		known[c.ID] = true
		check, e := core.NewCheck(*c, storage.NewMemoryList(100), prefs)
		if e != nil {
			return nil, fmt.Errorf("check %q: %w", c.Name, e)
		}
		if e = check.AddNotifiers(s.service, c.SendAlerts); e != nil {
			return nil, fmt.Errorf("check %q: %w", c.Name, e)
		}
		check.Notifiers = nil // AddChecks resolves and attaches them after persistence.
		prepared = append(prepared, check)
		notifications = append(notifications, c.SendAlerts)
		ids = append(ids, c.ID)
	}
	if err = appendStore.AppendChecks(batch); err != nil {
		return nil, fmt.Errorf("could not persist checks: %w", err)
	}
	ranks := make([]int, len(prepared))
	for i := range ranks {
		ranks[i] = len(current) + i
	}
	if err = s.service.AddChecks(prepared, notifications, ranks); err != nil {
		return nil, fmt.Errorf("checks persisted but activation failed; inspect check list before retrying: %w", err)
	}
	for _, check := range prepared {
		go check.Start()
	}
	return &pb.CheckAddResponse{Ids: ids}, nil
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

func Run(service *core.Service, port int, stores ...config.Store) {

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
	var store config.Store
	if len(stores) > 0 {
		store = stores[0]
	}
	filename := ""
	if fs, ok := store.(interface{ Filename() string }); ok {
		filename = fs.Filename()
	}
	pb.RegisterRedalertServiceServer(s, &server{service: service, store: store, filename: filename})
	s.Serve(lis)
}
