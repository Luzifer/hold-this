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
		CORS           bool   `flag:"cors" default:"false" description:"Add allow-all CORS headers for JS access"`
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

	if cfg.CORS {
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("origin"))
				w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")

				next.ServeHTTP(w, r)
			})
		})

		router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	}

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
