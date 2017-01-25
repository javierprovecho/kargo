package main

import (
	"fmt"

	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/kardianos/osext"
)

func main() {
	endpoint := "tcp://localhost:2375"
	client, _ := docker.NewClient(endpoint)
	swarm, _ := client.ListNodes(docker.ListNodesOptions{})
	fmt.Println(swarm)
	fmt.Println(osext.Executable())
	fmt.Println(os.Getwd())
}
