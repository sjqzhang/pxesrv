package pxecore

import (
	"fmt"
	"net"
)

// A Server boots machines using a Booter.
type Server struct {
	Config Config
	errs   chan error
}

var log = GetLogger("pxecore")

func (s *Server) Prepare() error {
	if err := s.LoadTemplates(); err != nil {
		log.Errorf("error during template loading, error: %s", err)
		return err
	}
	if err := s.RenderFile(); err != nil {
		log.Errorf("error during template rendering, error: %s", err)
		return err
	}
	log.Print("templates rendering successful")
	return nil
}

func (s *Server) Serve() error {

	dhcp, err := net.ListenPacket("udp4", fmt.Sprintf("%s:%s", s.Config.DHCP.IP, s.Config.DHCP.Port))
	if err != nil {
		log.Errorf("start DHCP failed, %s", err)
		return err
	}

	a, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", s.Config.TFTP.IP, s.Config.TFTP.Port))
	if err != nil {
		log.Errorf("resolveUDP failed, %s", err)
		return err
	}
	tftp, err := net.ListenUDP("udp4", a)
	if err != nil {
		log.Errorf("start TFTP failed, %s", err)
		dhcp.Close()
		return err
	}

	http, err := net.Listen("tcp4", fmt.Sprintf("%s:%s", s.Config.HTTP.IP, s.Config.HTTP.Port))
	if err != nil {
		log.Errorf("start HTTP failed, %s", err)
		dhcp.Close()
		tftp.Close()
		return err
	}
	// 4 buffer slots, one for each goroutine, plus one for
	// Shutdown(). We only ever pull the first error out, but shutdown
	// will likely generate some spurious errors from the other
	// goroutines, and we want them to be able to dump them without
	// blocking.
	s.errs = make(chan error, 5)

	//log.debug("Init", "Starting Pixiecore goroutines")

	go func() { s.errs <- s.serveDHCP(dhcp) }()
	go func() { s.errs <- s.serveTFTP(tftp) }()
	go func() { s.errs <- s.serveHTTP(http) }()

	// Wait for either a fatal error, or Shutdown().
	err = <-s.errs
	dhcp.Close()
	tftp.Close()
	http.Close()
	return err
}

// Shutdown causes Serve() to exit, cleaning up behind itself.
func (s *Server) Shutdown() {
	select {
	case s.errs <- nil:
	default:
	}
}
