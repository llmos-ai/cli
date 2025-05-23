package logserver

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	DefaultSocketLocation = "\x00logserver" // \x00 is the null byte, which we need to use for abstract namespace sockets instead of @ for some reason
)

// Server structure is used to the store backend information
type Server struct {
	SocketLocation string
	Debug          bool
}

// StartServerWithDefaults starts the server with default values. If the LLMOS_LOG_LEVEL environment variable is set,
// it will be parsed and used to set the log level.
func StartServerWithDefaults() {
	if level, err := logrus.ParseLevel(os.Getenv("LLMOS_LOG_LEVEL")); err == nil {
		logrus.SetLevel(level)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	s := Server{
		SocketLocation: DefaultSocketLocation,
	}
	s.Start()
}

// Start the server
func (s *Server) Start() {
	_ = os.Remove(s.SocketLocation)
	go func() {
		_ = s.ListenAndServe()
	}()
}

// ListenAndServe is used to setup handlers and
// start listening on the specified location
func (s *Server) ListenAndServe() error {
	server := http.Server{}
	http.HandleFunc("/v1/loglevel", s.loglevel)
	socketListener, err := net.Listen("unix", s.SocketLocation)
	if err != nil {
		logrus.Errorf("Failed to start logserver: %v", err)
		return err
	}
	return server.Serve(socketListener)
}

func (s *Server) loglevel(rw http.ResponseWriter, req *http.Request) {
	// curl -X POST -d "level=debug" localhost:12345/v1/loglevel
	logrus.Debugf("Received loglevel request")
	if req.Method == http.MethodGet {
		level := logrus.GetLevel().String()
		_, _ = rw.Write([]byte(fmt.Sprintf("%s\n", level)))
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("Failed to parse form: %v\n", err)))
		}
		level, err := logrus.ParseLevel(req.Form.Get("level"))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(fmt.Sprintf("Failed to parse loglevel: %v\n", err)))
		} else {
			logrus.SetLevel(level)
			_, _ = rw.Write([]byte("OK\n"))
		}
	}
}
