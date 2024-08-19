package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	httpHelper "github.com/Luzifer/go_helpers/v2/http"
	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		Gzip           bool   `flag:"gzip" default:"true" description:"Enable gzip compression"`
		Listen         string `flag:"listen" default:":3000" description:"Port/IP to listen on"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		RootDir        string `flag:"root-dir,r" default:"" vardefault:"rootDir" description:"Where to store files / get files from"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func initApp() error {
	rconfig.SetVariableDefaults(map[string]string{
		"rootDir": mustMkdirTemp(),
	})
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		return errors.Wrap(err, "parsing cli options")
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return errors.Wrap(err, "parsing log-level")
	}
	logrus.SetLevel(l)

	return nil
}

func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		logrus.WithField("version", version).Info("hold-this")
		os.Exit(0)
	}

	router := mux.NewRouter()
	router.PathPrefix("/").Methods(http.MethodDelete).HandlerFunc(handleDelete)
	router.PathPrefix("/").Methods(http.MethodGet).HandlerFunc(handleGet)
	router.PathPrefix("/").Methods(http.MethodPost, http.MethodPut).HandlerFunc(handlePut)

	var hdl http.Handler = router
	if cfg.Gzip {
		hdl = httpHelper.GzipHandler(hdl)
	}
	hdl = httpHelper.NewHTTPLogHandlerWithLogger(hdl, logrus.StandardLogger())

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           hdl,
		ReadHeaderTimeout: time.Second,
	}

	logrus.WithField("version", version).WithField("root_dir", cfg.RootDir).Info("hold-this started")
	if err = server.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("listening for HTTP traffic")
	}
}

func mustMkdirTemp() string {
	td, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	return td
}
