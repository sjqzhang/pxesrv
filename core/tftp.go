package core

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pin/tftp"
)

// readHandler is called when client starts file download from server
func (s *Service) tftpReadHandler(filename string, rf io.ReaderFrom) error {
	rootPath := filepath.Join(s.DocRoot, s.TFTPRoot, filename)
	file, err := os.Open(rootPath)
	if err != nil {
		log.Errorf("tftp err1: %v", err)
		return err
	}
	log.Infof("[TFTP]")
	n, err := rf.ReadFrom(file)
	if err != nil {
		log.Errorf("tftp err2: %v", err)
		return err
	}
	log.Infof("[TFTP] tftp_files %s(%d) bytes sent", filename, n)
	return nil
}

// writeHandler is called when client starts file upload to server
func (s *Service) tftWriteHandler(filename string, wt io.WriterTo) error {
	rootPath := filepath.Join(s.DocRoot, s.TFTPRoot, filename)
	file, err := os.OpenFile(rootPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}
	log.Infof("TFTP: tftp_files %d bytes received", n)
	log.Infof("TFTP: tftp_files recieved and stored file to %s", rootPath)
	return nil
}

func (s *Service) serveTFTP(l *net.UDPConn) error {
	rootPath := filepath.Join(s.DocRoot, s.TFTPRoot)
	tftpServer := tftp.NewServer(s.tftpReadHandler, s.tftWriteHandler)
	tftpServer.SetTimeout(5 * time.Second) // optional
	log.Infof("starting tftp server and listening on port %s handle on path: %s", s.TFTPPort, rootPath)
	tftpServer.Serve(l) // blocks until s.Shutdown() is called
	return nil
}