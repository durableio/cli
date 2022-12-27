package version

import (
	"fmt"
)

// Set during build
var (
	version   = "0.0.0"
	buildTime = "20221227"
)

func GetVersion() string {
	return fmt.Sprintf("durable %s [build-%s]", version, buildTime)
}
