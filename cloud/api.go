package cloud

import (
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/viper"
)

// A Service describes how to interact in a general way with the
// supported cloud services.
type Service interface {
	Register(device string) error
}

// GetService retrieves a service with a given name.
func GetService(name string) (Service, error) {
	if name == "azure-cli" {
		return &azureService{}, nil
	} else if name == "none" {
		return &emptyService{}, nil
	} else {
		return nil, errors.New("Unsupported service")
	}
}

// The azureService registers devices using the 'az' command line
// tool.
type azureService struct{}

// An empty cloud service that doesn't connect at all.
type emptyService struct{}

func (s *azureService) Register(device string) error {
	hubName := viper.GetString("server.hubname")
	resourceGroup := viper.GetString("server.resourcegroup")

	cmd := exec.Command("az",
		"iot", "hub", "device-identity", "create",
		"--device-id", device,
		"--auth-method", "x509_ca",
		"--resource-group", resourceGroup,
		"--hub-name", hubName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *emptyService) Register(device string) error {
	return nil
}
