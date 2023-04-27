package cmds

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"donuts/dc_logs/ui"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs for docker-compose",
	Run: func(cmd *cobra.Command, args []string) {
		ui.CreateUI()
		// RunLogs()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func RunLogs() {
	// Get where we are
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println("Work Dir:", pwd)

	// find docker-compose file
	dockerComposePath := pwd + "/docker-compose.yml"
	projectDir := path.Base(pwd)
	if _, err := os.Stat(dockerComposePath); errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist
		fmt.Println("ERR: Cannot Find Docker Compose file")
		panic(err)
	}
	fmt.Println("Docker Compose File Found:", dockerComposePath)
	fmt.Println("Project Directory:", projectDir)

	// find all services
	services := get_services(dockerComposePath)
	for index, val := range services {
		services[index] = projectDir + "-" + val
	}
	fmt.Println("Services:", services)

	// get container names
	containerNames := getContainerNames()

	// find the services we want logs for
	servicesToFollow := make(map[string][]string)
	for name, containerID := range containerNames {
		for _, service := range services {
			if strings.Contains(name, service) {
				fmt.Println("Service Found:", service, "ContainerName:", name, "ID:", containerID[:10])
				_, ok := servicesToFollow[service]
				if !ok {
					servicesToFollow[service] = make([]string, 0)
				}
				servicesToFollow[service] = append(servicesToFollow[service], containerID)
			}
		}
	}
	fmt.Println(servicesToFollow)

	// for each service start a goroutine
	log := make(chan string, 1000)
	for service, containers := range servicesToFollow {
		for _, CID := range containers {
			go dumpLogToChannel(log, service, CID)
		}
	}

	time.Sleep(300 * time.Second)

	// return channels for each service
}

func dumpLogToChannel(log_chan chan string, service string, containerID string) {
	// Connect to docker
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// get container log
	log, err := cli.ContainerLogs(
		context.Background(),
		containerID,
		types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: false,
			Follow:     true,
			Tail:       "100",
		})
	if err != nil {
		panic(err)
	}

	// forever dump container log to channel
	hdr := make([]byte, 8)
	for {
		_, err := log.Read(hdr)
		if err != nil {
			panic(err)
		}
		mess := service + ": "

		switch hdr[0] {
		case 1:
			mess += "STDOUT "
		default:
			mess += "STDERR "
		}

		count := binary.BigEndian.Uint32(hdr[4:])
		data := make([]byte, count)
		_, err = log.Read(data)
		if err != nil {
			panic(err)
		}

		mess += string(data)

		fmt.Println(mess)
	}
}

func get_services(dockerComposePath string) []string {
	viper_client := viper.New()
	viper_client.SetConfigType("yaml")
	ofile, err := os.Open(dockerComposePath)
	if err != nil {
		panic(err)
	}
	viper_client.ReadConfig(ofile)

	serviceMap := viper_client.GetStringMap("services")
	serviceArr := make([]string, 0)
	for key := range serviceMap {
		fmt.Println("Key:", key)
		serviceArr = append(serviceArr, key)
	}

	return serviceArr
}

func getContainerNames() map[string]string {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	containerNames := make(map[string]string)
	for _, container := range containers {
		fmt.Printf("%s %s %s %s %s\n", container.ID[:10], container.Image, container.Names, container.State, container.Status)

		for _, name := range container.Names {
			containerNames[name] = container.ID
		}
	}
	return containerNames
}
