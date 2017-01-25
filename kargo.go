package kargo

import (
	"flag"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kardianos/osext"
)

var (
	replicas int

	EnableKubernetes bool
	master           string

	EnableSwarm bool
	api         string

	deleteDeployment bool
)

func init() {
	flag.BoolVar(&EnableKubernetes, "kubernetes", false, "Deploy to Kubernetes.")
	flag.StringVar(&master, "master", "127.0.0.1:8001", "Kubernetes Master")

	flag.BoolVar(&EnableSwarm, "swarm", false, "Deploy to Swarm.")
	flag.StringVar(&api, "api", "tcp://127.0.0.1:2375", "Swarm API")

	flag.BoolVar(&deleteDeployment, "delete", false, "Delete current deployment (if any)")

	flag.IntVar(&replicas, "replicas", 1, "Number of replicas")

	flag.Parse()
}

type DeploymentConfig struct {
	Args      []string
	Env       map[string]string
	BinaryURL string
	Name      string
	Replicas  int
}

type DeploymentManager struct {
	DC DeploymentConfig
	UC UploadConfig
	R  func()
}

func (dm *DeploymentManager) Start() {

	if !EnableKubernetes && !EnableSwarm {
		dm.R()
		return
	}

	if EnableKubernetes && EnableSwarm {
		logrus.Fatalln("You must enable only one orchestrator.")
	}

	if deleteDeployment {

		if EnableKubernetes {
			err := deleteReplicaSet(dm.DC)
			if err != nil {
				logrus.Error(err)
			}
			return
		}

		if EnableSwarm {
			err := deleteService(dm.DC)
			if err != nil {
				logrus.Error(err)
			}
			return
		}
	}

	// Source or binary
	dm.UC.path, _ = osext.Executable()
	if strings.Contains(dm.UC.path, "go-build") {
		dm.UC.path = ""
	}

	// Upload (and build) or cached
	link, err := Upload(dm.UC)
	if err != nil {
		logrus.Fatalf("Error while uploading: %s\n", err)
	}

	dm.DC.BinaryURL = link

	if dm.DC.Env == nil {
		dm.DC.Env = make(map[string]string)
	}

	if EnableKubernetes {
		err = createReplicaSet(dm.DC)
		if err != nil {
			logrus.Error(err)
		}

	}
	if EnableSwarm {
		err = createService(dm.DC)
		if err != nil {
			logrus.Error(err)
		}
	}

}
