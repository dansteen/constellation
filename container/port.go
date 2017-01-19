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
	portArray[1] = fmt.Sprintf("%s:%d", port.Name, port.HostPort)
	return portArray
}

// SetHostPort will get a free port on the host machine and save it as the mapped port.  You want to do this as close to the actual
// running of the command as possible to avoid potential conflicts
func (port *Port) SetHostPort() error {
	// get an open port
	if strings.HasPrefix(port.Protocol, "tcp") {
		addr, err := net.ResolveTCPAddr(port.Protocol, "0.0.0.0:0")
		if err != nil {
			return err
		}
		conn, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		port.HostPort = conn.Addr().(*net.TCPAddr).Port
	} else if strings.HasPrefix(port.Protocol, "udp") {
		addr, err := net.ResolveUDPAddr(port.Protocol, "0.0.0.0:0")
		if err != nil {
			return err
		}
		conn, err := net.ListenUDP("upd", addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		port.HostPort = conn.LocalAddr().(*net.UDPAddr).Port
	}
	return nil
}
