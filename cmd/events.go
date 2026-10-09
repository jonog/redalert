package cmd

import (
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

func init() { RootCmd.AddCommand(eventsCmd) }
