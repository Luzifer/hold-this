package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const dirCreateMode = 0o750

func deriveFile(r *http.Request) (string, error) {
	absRoot, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return "", errors.Wrap(err, "getting absolute root")
	}

	file := filepath.Clean(path.Join(absRoot, strings.TrimLeft(r.URL.Path, "/")))
	if !strings.HasPrefix(file, path.Join(absRoot, "")) {
		return "", errors.New("break-out attempt")
	}

	return file, nil
}

func genericHTTPError(w http.ResponseWriter, reqID string, err error, desc string) {
	logrus.WithField("req_id", reqID).WithError(err).Error(desc)
	http.Error(w, fmt.Sprintf("something went wrong: %s", reqID), http.StatusInternalServerError)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.Must(uuid.NewV4()).String()

	f, err := deriveFile(r)
	if err != nil {
		genericHTTPError(w, reqID, err, "deriving file for request")
		return
	}

	stat, err := os.Stat(f)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		genericHTTPError(w, reqID, err, "stating file")
		return
	}

	if stat.IsDir() {
		genericHTTPError(w, reqID, errors.New("is directory"), "refusing to delete dir")
		return
	}

	if err = os.Remove(f); err != nil {
		genericHTTPError(w, reqID, err, "deleting file")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.Must(uuid.NewV4()).String()

	f, err := deriveFile(r)
	if err != nil {
		genericHTTPError(w, reqID, err, "deriving file for request")
		return
	}

	stat, err := os.Stat(f)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		genericHTTPError(w, reqID, err, "stating file")
		return
	}

	if stat.IsDir() {
		genericHTTPError(w, reqID, errors.New("is directory"), "refusing to get dir")
		return
	}

	http.ServeFile(w, r, f)
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.Must(uuid.NewV4()).String()

	f, err := deriveFile(r)
	if err != nil {
		genericHTTPError(w, reqID, err, "deriving file for request")
		return
	}

	if err = os.MkdirAll(path.Dir(f), dirCreateMode); err != nil {
		genericHTTPError(w, reqID, err, "ensuring directory for file")
		return
	}

	out, err := os.Create(f)
	if err != nil {
		genericHTTPError(w, reqID, err, "creating file")
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, r.Body); err != nil {
		genericHTTPError(w, reqID, err, "copying file contents")
		return
	}

	w.WriteHeader(http.StatusCreated)
}
