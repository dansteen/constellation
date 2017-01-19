package container

import (
	"fmt"
	"net"
	"strings"

	"github.com/dansteen/constellation/rkt"
)

// Port represents a port that is defined in a container manifest.  We ingest all the values even though we only use a few.
type Port struct {
	rkt.ImageAppPort
	HostPort int
}

// GenerateCommandLine will generate the command line options to use this port
func (port *Port) GenerateCommandLine() []string {
	portArray := make([]string, 2)
	portArray[0] = "--port"
	portArray[1] = fmt.Sprintf("%s:%s", port.Name, port.HostPort)
	return portArray
}

// SetHostPort will get a free port on the host machine and save it as the mapped port.  You want to do this as close to the actual
// running of the command as possible to avoid potential conflicts
func (port *Port) SetHostPort() error {
	// get an open port
	if strings.HasPrefix(port.Protocol, "tcp") {
		addr, err := net.ResolveTCPAddr(port.Protocol, "localhost:0")
		if err != nil {
			return err
		}
		port.HostPort = addr.Port
	} else if strings.HasPrefix(port.Protocol, "udp") {
		addr, err := net.ResolveUDPAddr(port.Protocol, "localhost:0")
		if err != nil {
			return err
		}
		port.HostPort = addr.Port
	}
	return nil
}
