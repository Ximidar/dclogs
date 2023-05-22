package cmds

import (
	"fmt"
	"log"
	"os"
	"sort"
	getlogs "ximidar/dc_logs/getLogs"
	"ximidar/dc_logs/logs"
	"ximidar/dc_logs/ui"

	"github.com/spf13/cobra"
)

var Log *log.Logger

var rootCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs for docker-compose",
	Run: func(cmd *cobra.Command, args []string) {

		// Get where we are
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		// set up logs
		logPath := pwd + "/dc_logs.log"
		file, err := os.Create(logPath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		logs.Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
		logs.Log.Println("Starting DC_LOGS")

		// TODO take out
		pwd += "/compose_test"

		// create logs backend
		logs := getlogs.NewGetLogs()
		logs.InitAtPath(pwd)

		// create ui
		ui := ui.CreateUI(logs)

		// populate ui
		ui.LogSelector.SetRoot(logs.ProjectName)
		for _, service := range logs.Services {
			containers := logs.Containers[service]
			names := make([]string, 0)
			for _, container := range containers {
				names = append(names, container.Name)
			}
			// Sort names
			sort.Strings(names)
			// push node
			ui.LogSelector.AddNode(service, names...)

		}

		ui.Start()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
