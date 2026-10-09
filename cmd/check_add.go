package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	pb "github.com/jonog/redalert/servicepb"
	"github.com/spf13/cobra"
	netcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var checkAddInput string

var checkAddCmd = &cobra.Command{
	Use:   "check-add",
	Short: "Add checks from a JSON array and start monitoring them",
	RunE: func(cmd *cobra.Command, args []string) error {
		var input io.Reader
		if checkAddInput == "-" {
			input = os.Stdin
		} else {
			f, err := os.Open(checkAddInput)
			if err != nil {
				return fmt.Errorf("open input: %w", err)
			}
			defer f.Close()
			input = f
		}
		data, err := io.ReadAll(io.LimitReader(input, 4<<20))
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		connectCtx, connectCancel := netcontext.WithTimeout(netcontext.Background(), 30*time.Second)
		defer connectCancel()
		conn, err := grpc.DialContext(connectCtx, fmt.Sprintf("localhost:%d", rpcPort), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("connect to server: %w", err)
		}
		defer conn.Close()
		ctx, cancel := netcontext.WithTimeout(netcontext.Background(), 30*time.Second)
		defer cancel()
		response, err := pb.NewRedalertServiceClient(conn).CheckAdd(ctx, &pb.CheckAddRequest{Destination: cmd.Flag("config-file").Value.String(), Json: string(data)})
		if err != nil {
			return fmt.Errorf("add checks failed (if the server reports persistence uncertainty, inspect check-list before retrying): %w", err)
		}
		for _, id := range response.Ids {
			fmt.Fprintln(cmd.OutOrStdout(), id)
		}
		return nil
	},
}

func init() {
	checkAddCmd.Flags().StringVarP(&checkAddInput, "input", "i", "", "JSON array file, or - for stdin")
	_ = checkAddCmd.MarkFlagRequired("input")
	RootCmd.AddCommand(checkAddCmd)
}
