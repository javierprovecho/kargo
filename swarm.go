package kargo

import (
	"context"

	"time"

	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
)

func createService(config DeploymentConfig) error {

	client, err := dockerClient.NewClient(api, "1.12.6", http.DefaultClient, nil)
	if err != nil {
		return err
	}

	env := []string{"URL=" + config.BinaryURL}

	if len(config.Env) > 0 {
		for name, value := range config.Env {
			env = append(env, name+"="+value)
		}
	}

	delay := time.Duration(time.Second)
	replicas := uint64(config.Replicas)

	svc := swarm.ServiceSpec{
		swarm.Annotations{
			config.Name,
			nil,
		},
		swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image: "javierprovecho/kargo-alpine",
				Args:  config.Args,
				Env:   env,
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition: swarm.RestartPolicyConditionAny,
				Delay:     &delay,
			},
		},
		swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		nil,
		nil,
		&swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				swarm.PortConfig{
					Name:          "web",
					Protocol:      swarm.PortConfigProtocolTCP,
					PublishedPort: 8081,
					TargetPort:    8081,
				},
			},
		},
	}

	opts := types.ServiceCreateOptions{}

	client.ServiceCreate(context.TODO(), svc, opts)
	if err != nil {
		return err
	}

	return nil
}

func deleteService(config DeploymentConfig) error {

	client, err := dockerClient.NewClient(api, "1.12.6", http.DefaultClient, nil)
	if err != nil {
		return err
	}

	client.ServiceRemove(context.TODO(), config.Name)
	if err != nil {
		return err
	}

	return nil
}
