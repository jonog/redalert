package cmd

import (
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"
	pb "github.com/jonog/redalert/servicepb"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// checksCmd represents the checks command
var checksCmd = &cobra.Command{
	Use:   "checks",
	Short: "List checks",
	Long:  "List checks",
	Run: func(cmd *cobra.Command, args []string) {

		conn, err := grpc.Dial("localhost:"+strconv.Itoa(rpcPort), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewRedalertServiceClient(conn)

		r, err := c.CheckList(context.Background(), &pb.CheckListRequest{})
		if err != nil {
			log.Fatalf("could not get response: %v", err)
		}

		data := [][]string{}
		for _, check := range r.Members {
			data = append(data, []string{check.ID, check.Name, colorStatus(check.Status)})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Status"})
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.AppendBulk(data)
		table.Render()
	},
}

func init() {
	RootCmd.AddCommand(checksCmd)
}

func colorStatus(status pb.Check_Status) string {
	switch status {
	case pb.Check_DISABLED:
		return color.WhiteString(status.String())
	case pb.Check_UNKNOWN:
		return color.YellowString(status.String())
	case pb.Check_FAILING:
		return color.RedString(status.String())
	case pb.Check_SUCCESSFUL:
		return color.GreenString(status.String())
	}
	return "-"
}
