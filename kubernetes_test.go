package kubernetes

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/jsdir/deployer/pkg/resources"
	"github.com/stretchr/testify/assert"
)

type Pod struct {
	Labels map[string]string
}

type PodList struct {
	Items []Pod
}

func deploy(d *resources.Deploy) error {
	k := new(Kubernetes)
	return k.Deploy(d)
}

func TestErrorsHandled(t *testing.T) {
	err := deploy(&resources.Deploy{
		EnvConfig: map[string]string{
			"Cmd":          "invalid_cmd",
			"ManifestGlob": "./test/*.json",
		},
	})
	assert.EqualError(t, err, "Failed to run command")
}

func TestClusterState(t *testing.T) {
	// First deploy
	err := deploy(&resources.Deploy{
		Env: &resources.Environment{
			ReleaseId:    1,
			Updated:      "time",
			DeployActive: true,
		},
		LastRelease: &resources.Release{
			Id:   1,
			Name: "release-1",
			Config: map[string]string{
				"config_key": "config_value",
			},
			Services: map[string]string{
				"service_1": "1",
				"service_2": "1",
			},
		},
		Release: &resources.Release{
			Id:   2,
			Name: "release-2",
			Config: map[string]string{
				"config_key": "config_value",
			},
			Services: map[string]string{
				"service_1": "2",
				"service_2": "1",
			},
		},
		ChangedServices: []string{"service_1"},
		EnvConfig: map[string]string{
			"Cmd":          "kubectl --server=http://localhost:8888",
			"ManifestGlob": "./test/*.json",
		},
	})

	assert.NoError(t, err)

	// Check that the cluster state was changed.
	pods, err := exec.Command(
		"kubectl",
		"--server=http://localhost:8888",
		"--output=json",
		"get",
		"pods",
	).Output()
	assert.NoError(t, err)

	// Decode the resposne
	podList := PodList{}
	err = json.Unmarshal(pods, &podList)
	assert.NoError(t, err)

	for i := 0; i < 2; i++ {
		labels := podList.Items[i].Labels
		assert.Equal(t, labels["changed"], "[service_1]")
	}
}
