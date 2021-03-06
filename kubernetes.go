package kubernetes

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/jsdir/deployer/pkg/resources"

	"github.com/mitchellh/mapstructure"
)

type KubernetesConfig struct {
	ManifestGlob string
	Cmd          string
}

type Kubernetes struct{}

func (k *Kubernetes) Deploy(deploy *resources.Deploy) error {
	// Get environment config.
	config := KubernetesConfig{
		ManifestGlob: "./manifests/*.json",
		Cmd:          "kubectl",
	}
	err := mapstructure.Decode(deploy.EnvConfig, &config)
	if err != nil {
		return err
	}

	log.Println(config)

	// Iterate through and parse templates.
	templates, err := filepath.Glob(config.ManifestGlob)
	if err != nil {
		return err
	}

	// Execute commands in parallel.
	count := len(templates)
	if count == 0 {
		log.Println("Info: No templates found")
		return nil
	}

	sem := make(chan error, count)
	for _, filename := range templates {
		go func(filename string) {
			sem <- updateManifest(filename, deploy, &config)
			return
		}(filename)
	}

	for i := 0; i < count; i++ {
		err := <-sem
		if err != nil {
			log.Println(err)
			return errors.New("Failed to run command")
		}
	}

	return nil
}

func updateManifest(filename string, deploy *resources.Deploy, config *KubernetesConfig) error {
	log.Printf("Uploading %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	tmpl, err := template.New("").Parse(string(data))
	if err != nil {
		return err
	}

	runCmd := func(action string) ([]byte, error) {
		log.Println("Running command:", config.Cmd, action)
		args := append(strings.Split(config.Cmd, " "), action, "-f", "-")
		cmd := exec.Command(args[0], args[1:]...)

		// Attach pipes.
		cmd.Stdout = os.Stdout
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}

		// Start command.
		err = cmd.Start()
		if err != nil {
			return nil, err
		}

		err = tmpl.Execute(stdin, deploy)
		if err != nil {
			return nil, err
		}

		stdin.Close()
		stderrData, err := ioutil.ReadAll(stderr)
		if err != nil {
			return nil, err
		}

		return stderrData, cmd.Wait()
	}

	// Run commands.
	stderr, err := runCmd("create")

	// Try updating if the resource already exists.
	if err != nil {
		matched, matchErr := regexp.Match("already exists", stderr)
		if matchErr == nil && matched {
			// Temporarily delete before create until ReplicationControllers are supported.
			deleteStderr, deleteErr := runCmd("delete")
			if deleteErr != nil {
				fmt.Println(string(deleteStderr[:]))
				return deleteErr
			}
			createStderr, createErr := runCmd("create")
			if createErr != nil {
				fmt.Println(string(createStderr[:]))
				return createErr
			}
			return nil
		}

		fmt.Println(string(stderr[:]))
		return err
	}

	return nil
}
