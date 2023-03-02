package apptest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	filedriver "github.com/goftp/file-driver"
	"github.com/goftp/server"
	log "github.com/sirupsen/logrus"
)

func StartHTTPServer(isHTTPS bool) (u *url.URL, srv *httptest.Server, err error) {
	if isHTTPS {
		srv = httptest.NewTLSServer(http.FileServer(http.Dir(".")))
	} else {
		srv = httptest.NewServer(http.FileServer(http.Dir(".")))
	}

	u, err = url.Parse(srv.URL)

	return
}

func StartFTPServer(ctx context.Context, port int, waitForStarting time.Duration) (*url.URL, *server.Server, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	ftpHost := fmt.Sprintf("ftp://root:root@localhost:%d", port)
	ftpHostURL, err := url.Parse(ftpHost)
	if err != nil {
		return nil, nil, err
	}

	factory := &filedriver.FileDriverFactory{
		RootPath: path,
		Perm:     server.NewSimplePerm("user", "group"),
	}

	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     port,
		Hostname: "localhost",
		Auth:     &server.SimpleAuth{Name: "root", Password: "root"},
	}

	log.Printf("Starting ftp server on %v:%v", opts.Hostname, opts.Port)
	ftpSrvr := server.NewServer(opts)

	go func() {
		<-ctx.Done()
		err := ftpSrvr.Shutdown()
		if err != nil {
			log.Error(err)
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		err := ftpSrvr.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return ftpHostURL, nil, err
	case <-time.After(waitForStarting):
		return ftpHostURL, ftpSrvr, nil
	}
}
