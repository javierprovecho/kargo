package main

import (
	"github.com/gin-gonic/gin"
	"github.com/javierprovecho/kargo"
)

func main() {
	deployment := &kargo.DeploymentManager{
		DC: kargo.DeploymentConfig{
			Args:     []string{"-p", "8082"},
			Env:      map[string]string{"PORT": "8081"},
			Name:     "gomad",
			Replicas: 2,
		},
		UC: kargo.UploadConfig{
			BucketName: "mycorp-apps",
			ObjectName: "gomad",
			ProjectID:  "robust-index-125619",
		},
		R: server,
	}

	deployment.Start()
}

func server() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.Run()
}
