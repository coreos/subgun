package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/coreos/go-systemd/activation"
)

func ServeTCP(handle http.Handler, port string) {
	http.ListenAndServe(":"+port, handle)
}

func ServeFD(handle http.Handler) error {
	println("Looking for listeners")
	ls, e := listenFD()
	fmt.Printf("Got %d listeners", len(ls))
	if e != nil {
		println(e.Error())
	}

	chErrors := make(chan error, len(ls))

	// Since listenFD will return one or more sockets we have
	// to create a go func to spawn off multiple serves
	for i, _ := range ls {
		listener := ls[i]
		go func() {
			httpSrv := http.Server{Handler: handle}
			chErrors <- httpSrv.Serve(listener)
		}()
	}

	for i := 0; i < len(ls); i += 1 {
		err := <-chErrors
		if err != nil {
			return err
		}
	}

	return nil
}

func listenFD() ([]net.Listener, error) {
	files := activation.Files(false)
	if files == nil || len(files) == 0 {
		return nil, errors.New("No sockets found")
	}

	listeners := make([]net.Listener, len(files))
	for i, f := range files {
		var err error
		listeners[i], err = net.FileListener(f)
		if err != nil {
			return nil, fmt.Errorf("Error setting up FileListener for fd %d: %s", f.Fd(), err.Error())
		}
	}

	return listeners, nil
}
