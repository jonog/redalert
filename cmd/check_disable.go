package cmd

import (
	"log"
	"strconv"

	pb "github.com/jonog/redalert/servicepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"
)

// checkDisableCmd represents the check-disable command
var checkDisableCmd = &cobra.Command{
	Use:   "check-disable",
	Short: "Disables a check",
	Long:  "Disables a check",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			log.Fatalf("not enough args")
		}

		conn, err := grpc.Dial("localhost:"+strconv.Itoa(rpcPort), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		c := pb.NewRedalertServiceClient(conn)
		_, err = c.CheckDisable(context.Background(), &pb.CheckDisableRequest{ID: args[0]})
		if err != nil {
			log.Fatalf("could not get response: %v", err)
		}

	},
}

func init() {
	RootCmd.AddCommand(checkDisableCmd)
}
