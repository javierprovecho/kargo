package kargo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
)

var (
	replicasetsEndpoint = "/apis/extensions/v1beta1/namespaces/%s/replicasets"
	replicasetEndpoint  = "/apis/extensions/v1beta1/namespaces/%s/replicasets/%s"
	scaleEndpoint       = "/apis/extensions/v1beta1/namespaces/%s/replicasets/%s/scale"
	podsEndpoint        = "/api/v1/namespaces/%s/pods"
	ErrNotExist         = errors.New("does not exist")
)

func createReplicaSet(config DeploymentConfig) error {

	container := Container{
		Args:  config.Args,
		Image: "javierprovecho/kargo-alpine",
		Name:  config.Name,
		Env:   []EnvVar{EnvVar{Name: "URL", Value: config.BinaryURL}},
	}

	if len(config.Env) > 0 {
		for name, value := range config.Env {
			container.Env = append(container.Env, EnvVar{Name: name, Value: value})
		}
	}

	labels := map[string]string{"app": config.Name}

	rs := ReplicaSet{
		ApiVersion: "extensions/v1beta1",
		Kind:       "ReplicaSet",
		Metadata: Metadata{
			Name:      config.Name,
			Namespace: "default",
		},
		Spec: ReplicaSetSpec{
			Replicas: int64(config.Replicas),
			Selector: LabelSelector{
				MatchLabels: labels,
			},
			Template: PodTemplate{
				Metadata: Metadata{
					Labels: labels,
				},
				Spec: PodSpec{
					Containers: []Container{container},
				},
			},
		},
	}

	var b []byte
	body := bytes.NewBuffer(b)
	err := json.NewEncoder(body).Encode(rs)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(replicasetsEndpoint, "default")
	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPost,
		URL: &url.URL{
			Host:   master,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		logrus.Infoln(string(data))
		return errors.New("ReplicaSet: Unexpected HTTP status code" + resp.Status)
	}

	return nil
}

func deleteReplicaSet(config DeploymentConfig) error {
	err := scaleReplicaSet("default", config.Name, 0)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(replicasetEndpoint, "default", config.Name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodDelete,
		URL: &url.URL{
			Host:   master,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrNotExist
	}
	if resp.StatusCode != 200 {
		return errors.New("Delete ReplicaSet error non 200 reponse: " + resp.Status)
	}

	return nil
}

func getScale(namespace, name string) (*Scale, error) {
	var scale Scale

	path := fmt.Sprintf(scaleEndpoint, namespace, name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:   master,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get scale error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&scale)
	if err != nil {
		return nil, err
	}
	return &scale, nil
}

func scaleReplicaSet(namespace, name string, replicas int) error {
	scale, err := getScale(namespace, name)
	if err != nil {
		return err
	}
	scale.Spec.Replicas = int64(replicas)

	var b []byte
	body := bytes.NewBuffer(b)
	err = json.NewEncoder(body).Encode(scale)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(scaleEndpoint, namespace, name)
	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPut,
		URL: &url.URL{
			Host:   master,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrNotExist
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return errors.New("Scale ReplicaSet error non 200 reponse: " + resp.Status)
	}

	return nil
}
