package caserver

import (
	"log"
	"time"

	"github.com/Linaro/lite_bootstrap_server/cloud"
)

const interval = 10 * time.Second

// registration periodically checks the database for devices that have
// become known but have not been registered with the cloud service,
// and attempts to register them.
func registration() {
	cloud, err := cloud.GetService("azure-cli")
	if err != nil {
		log.Fatalf("Unable to get cloud service handle: %s\n", err)
	}

	for {
		time.Sleep(interval)

		devs, err := db.UnregisteredDevices()
		if err != nil {
			log.Printf("Warning: Unable to query db for devices: %s\n", err)
			continue
		}

		for _, dev := range devs {
			log.Printf("Register device: %s\n", dev)

			err = cloud.Register(dev)
			if err != nil {
				log.Printf("Warning: Unable to register device: %s\n", err)
				continue
			}

			err = db.MarkRegistered(dev)
			if err != nil {
				log.Printf("Warning: Unable to update database to register device: %s\n", err)
				continue
			}
		}
	}
}
