package cloud

import (
	"errors"
	"os"
	"os/exec"
)

type CloudService interface {
	Register(device string) error
}

// GetService retrieves a service with a given name.
func GetService(name string) (CloudService, error) {
	if name == "azure-cli" {
		return &azureService{}, nil
	} else {
		return nil, errors.New("Unsupported service")
	}
}

// The azureService registers devices using the 'az' command line
// tool.
type azureService struct {
}

var HubName string = "hubname"
var ResourceGroup string = "resourcegroup"

func (s *azureService) Register(device string) error {

	cmd := exec.Command("az",
		"iot", "hub", "device-identity", "create",
		"--device-id", device,
		"--auth-method", "x509_ca",
		"--resource-group", ResourceGroup,
		"--hub-name", HubName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
