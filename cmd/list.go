package cmd

import (
	"os"
	"strings"

	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:     "list",
		Short:   "List K3s clusters",
		Example: `  autok3s list`,
	}
)

func ListCommand() *cobra.Command {
	listCmd.Run = func(cmd *cobra.Command, args []string) {
		listCluster()
	}
	return listCmd
}

func listCluster() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Name", "Region", "Provider", "Status", "Masters", "Workers", "Version"})

	base := cluster.NewBaseProvider()
	filters, err := base.ListCluster()

	if err != nil {
		logrus.Fatalf("list cluster error %v \n", err)
	}

	for _, f := range filters {
		context := strings.Split(f.Name, ".")
		table.Append([]string{
			context[0],
			f.Region,
			f.Provider,
			f.Status,
			f.Master,
			f.Worker,
			f.Version,
		})
	}

	table.Render()
}
