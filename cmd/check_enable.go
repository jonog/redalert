package cmd

import (
	"log"
	"strconv"

	pb "github.com/jonog/redalert/servicepb"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// checkEnableCmd represents the check-enable command
var checkEnableCmd = &cobra.Command{
	Use:   "check-enable",
	Short: "Enables a check",
	Long:  "Enables a check",
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
		_, err = c.CheckEnable(context.Background(), &pb.CheckEnableRequest{ID: args[0]})
		if err != nil {
			log.Fatalf("could not get response: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(checkEnableCmd)
}
