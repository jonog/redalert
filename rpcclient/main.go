package main

import (
	"log"
	"os"

	"github.com/fatih/color"
	pb "github.com/jonog/redalert/service"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRedalertServiceClient(conn)

	r, err := c.ListChecks(context.Background(), &pb.ListChecksRequest{})
	if err != nil {
		log.Fatalf("could not get response: %v", err)
	}

	data := [][]string{}
	for _, check := range r.Members {
		data = append(data, []string{shortID(check.ID), check.Name, colorStatus(check.Status)})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Status"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
}

func colorStatus(status pb.Check_Status) string {
	switch status {
	case pb.Check_UNKNOWN:
		return color.YellowString(status.String())
	case pb.Check_FAILING:
		return color.RedString(status.String())
	case pb.Check_RECOVERED:
		return color.WhiteString(status.String())
	case pb.Check_NORMAL:
		return color.GreenString(status.String())
	}
	return "-"
}

func shortID(longID string) string {
	return longID[:6]
}
