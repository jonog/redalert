package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	pb "github.com/jonog/redalert/servicepb"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var eventsCmd = &cobra.Command{
	Use: "events <check-id>", Short: "List retained events for a check",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected exactly one check ID")
		}
		conn, err := grpc.Dial("localhost:"+strconv.Itoa(rpcPort), grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("connect to server: %w", err)
		}
		defer conn.Close()
		response, err := pb.NewRedalertServiceClient(conn).EventList(context.Background(), &pb.EventListRequest{ID: args[0]})
		if err != nil {
			return fmt.Errorf("list events: %w", err)
		}
		w := cmd.OutOrStdout()
		jsonOutput, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		if jsonOutput {
			events := make([]jsonEvent, 0, len(response.Events))
			for _, event := range response.Events {
				encoded := jsonEvent{Time: event.Time, Tags: event.Tags, Messages: event.Messages, Metadata: event.Metadata}
				for _, metric := range event.Metrics {
					encoded.Metrics = append(encoded.Metrics, jsonMetric{Name: metric.Name, Value: metric.Value, Present: metric.Present})
				}
				events = append(events, encoded)
			}
			if err := json.NewEncoder(w).Encode(events); err != nil {
				return fmt.Errorf("encode events as JSON: %w", err)
			}
			return nil
		}
		if len(response.Events) == 0 {
			_, err = fmt.Fprintln(w, "No events found.")
			return err
		}
		for i, event := range response.Events {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "Time: %s\n", event.Time)
			fmt.Fprintf(w, "Tags: %s\n", formatMap(event.Tags))
			fmt.Fprintf(w, "Messages: %s\n", strings.Join(event.Messages, "; "))
			metrics := make([]string, 0, len(event.Metrics))
			for _, metric := range event.Metrics {
				value := "null"
				if metric.Present {
					value = strconv.FormatFloat(metric.Value, 'g', -1, 64)
				}
				metrics = append(metrics, metric.Name+"="+value)
			}
			sort.Strings(metrics)
			fmt.Fprintf(w, "Metrics: %s\n", strings.Join(metrics, ", "))
			fmt.Fprintf(w, "Metadata: %s\n", formatMap(event.Metadata))
		}
		return nil
	},
}

type jsonEvent struct {
	Time     string            `json:"time,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Messages []string          `json:"messages,omitempty"`
	Metrics  []jsonMetric      `json:"metrics,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type jsonMetric struct {
	Name    string  `json:"name,omitempty"`
	Value   float64 `json:"value,omitempty"`
	Present bool    `json:"present"`
}

func formatMap(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+values[key])
	}
	return strings.Join(pairs, ", ")
}

func init() {
	eventsCmd.Flags().Bool("json", false, "output events as JSON")
	RootCmd.AddCommand(eventsCmd)
}
