package getlogs

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"ximidar/dc_logs/logs"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/viper"
)

type DContainer struct {
	Name      string
	Container types.Container
}

type GetLogs struct {
	ProjectName       string
	Services          []string
	Containers        map[string][]DContainer
	selectedContainer DContainer

	WriteLog   chan string
	exit       chan bool
	first      bool
	lock       sync.Mutex
	currentLog io.ReadCloser
}

func NewGetLogs() *GetLogs {
	gl := new(GetLogs)

	gl.Services = make([]string, 0)
	gl.Containers = make(map[string][]DContainer)
	gl.selectedContainer = DContainer{
		Name: "",
	}
	gl.WriteLog = make(chan string)
	gl.exit = make(chan bool, 1)
	gl.first = true

	return gl
}

func (gl *GetLogs) InitAtPath(dirpath string) {
	gl.getServices(dirpath)
	gl.getContainers()
}

func (gl *GetLogs) GetChannel() chan string {
	return gl.WriteLog
}

func (gl *GetLogs) SelectCallback(containerName string) {
	logs.Log.Println("Select Callback For:", containerName)
	if !strings.Contains(containerName, "/") {
		// there's no subcontainer
		containers, ok := gl.Containers[containerName]
		if !ok {
			logs.Log.Println("Err: Service not found ", containerName)
			panic("service not found")
		}
		if len(containers) >= 1 {
			return
		}
		gl.selectedContainer = containers[0]
	} else {
		ids := strings.Split(containerName, "/")
		service := ids[0]
		name := ids[1]

		containers, ok := gl.Containers[service]
		if !ok {
			logs.Log.Println("Err: Service not found ", containerName)
			panic("service not found")
		}

		for _, container := range containers {
			if container.Name == name {
				gl.selectedContainer = container
				break
			}
		}

	}

	logs.Log.Println("Selected container:", gl.selectedContainer.Name)
	if !gl.first {
		gl.exit <- true
		gl.currentLog.Close()
		gl.lock.Lock()
		defer gl.lock.Unlock()
	} else {
		gl.first = false
	}

	go gl.dumpLogToChannel()

}

// GetServices will get the available services at the supplied path
func (gl *GetLogs) getServices(dirPath string) {
	// find docker-compose file
	dockerComposePath := dirPath + "/docker-compose.yml"
	gl.ProjectName = path.Base(dirPath)
	if _, err := os.Stat(dockerComposePath); errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist
		logs.Log.Println("ERR: Cannot Find Docker Compose file")
		panic(err)
	}

	// find all services
	gl.Services = getServices(dockerComposePath)
}

func (gl *GetLogs) getContainers() {
	if gl.Services == nil {
		err := errors.New("services is not initialized")
		panic(err)
	}

	for _, service := range gl.Services {
		containers := getContainers(service)
		gl.Containers[service] = containers

	}

}

func (gl *GetLogs) dumpLogToChannel() {
	gl.lock.Lock()
	defer gl.lock.Unlock()
	if gl.selectedContainer.Name == "" {
		// Nothing selected
		return
	}
	resource_name := gl.selectedContainer.Name
	logs.Log.Println("Attempting to get", gl.selectedContainer.Name)
	logs.Log.Println("ID:", gl.selectedContainer.Container.ID)

	// Connect to docker
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// get container log
	log, err := cli.ContainerLogs(
		context.Background(),
		gl.selectedContainer.Container.ID,
		types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: false,
			Follow:     true,
			Tail:       "10",
		})
	if err != nil {
		panic(err)
	}
	gl.currentLog = log
	defer log.Close()

	// forever dump container log to channel
	hdr := make([]byte, 8)
	for {
		select {
		case exit := <-gl.exit:
			if exit {
				logs.Log.Println("Exiting due to being asked. Resource:", resource_name)
				return
			}
		default:
			bytes_read, err := log.Read(hdr)

			if bytes_read == 0 {
				continue
			}
			if err != nil {
				panic(err)
			}
			mess := ""

			// hrd[0] == 1 STDOUT
			// anything else is STDERROR
			if hdr[0] != 1 {
				mess += "STDERR: "
			}

			count := binary.BigEndian.Uint32(hdr[4:])
			data := make([]byte, count)
			_, err = log.Read(data)
			if err != nil {
				gl.WriteLog <- err.Error()
			}

			mess += string(data)
			gl.WriteLog <- mess
		}
	}
}

func getServices(dockerComposePath string) []string {
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
		logs.Log.Println("Key:", key)
		serviceArr = append(serviceArr, key)
	}

	return serviceArr
}

func getContainers(serviceName string) []DContainer {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	DContainers := make([]DContainer, 0)

	for _, container := range containers {
		logs.Log.Printf("%s %s %s %s %s\n", container.ID[:10], container.Image, container.Names, container.State, container.Status)
		name := container.Names[0][1:]

		// find clean name
		name_pos := strings.Index(name, serviceName)
		cleanName := name
		if name_pos >= 0 {
			cleanName = name[name_pos:]
		}
		logs.Log.Println("Found Container:", name, cleanName)
		if strings.Contains(name, serviceName) {
			dcont := DContainer{
				Name:      cleanName,
				Container: container,
			}
			DContainers = append(DContainers, dcont)
		}
	}
	return DContainers
}
