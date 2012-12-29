package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"
)

func StaticHandler(w http.ResponseWriter, req *http.Request) {
	var filename = path.Join(*StaticRoot, req.URL.Path)
	f, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		defer f.Close()
	}

	now := time.Now().UTC()
	d := time.Date(2011, 4, 11, 3, 0, 0, 0, time.UTC)
	TTL := int64(86400)
	ttl := TTL - (now.Unix()-d.Unix())%TTL // shift
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", ttl))

	http.ServeFile(w, req, filename)
}
