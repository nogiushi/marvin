package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func ListenAndServe(address string, scheduler *scheduler) {
	http.Handle("/", http.HandlerFunc(makePageHandler(scheduler)))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Print("ListenAndServe:", err)
	}
}

type page struct {
	Name      string
	Title     string
	NotFound  bool
	Scheduler *scheduler
	t         *template.Template
}

func newPage(title string) *page {
	return &page{Title: title}
}

func (p *page) template() *template.Template {
	if p.t == nil {
		if t, err := template.ParseFiles("templates/site.html", "templates/"+p.Name+".html"); err == nil {
			p.t = t
		} else {
			log.Fatal(err)
		}
	}
	return p.t
}

func (p *page) Write(w http.ResponseWriter, req *http.Request) (err error) {
	var bw bytes.Buffer
	h := md5.New()
	mw := io.MultiWriter(&bw, h)
	err = p.template().Execute(mw, p)
	if err == nil {
		w.Header().Set("ETag", fmt.Sprintf("\"%x\"", h.Sum(nil)))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", bw.Len()))
		w.Write(bw.Bytes())
	} else {
		log.Println("template error:", err)
	}

	return err
}

func NotFoundHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "max-age=10, must-revalidate")
	w.WriteHeader(http.StatusNotFound)
	page := newPage("Not Found")
	page.Name = "notfound"
	page.NotFound = true
	page.Write(w, req)
	return
}

func setCacheControl(w http.ResponseWriter, req *http.Request) {
	if req.Header["X-Draft"] != nil {
		w.Header().Set("Cache-Control", "max-age=1, must-revalidate")
	} else {
		now := time.Now().UTC()
		//d := time.Time{2011, 4, 11, 3, 0, 0, 0, time.Monday, 0, "UTC"}
		d := time.Date(2011, 4, 11, 3, 0, 0, 0, time.UTC)
		TTL := int64(86400)
		ttl := TTL - (now.Unix()-d.Unix())%TTL // shift
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", ttl))
	}
}

func makePageHandler(s *scheduler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			if req.Method == "POST" {
				if err := req.ParseForm(); err == nil {
					name := req.Form["do_transition"]
					s.Hue.Do(name[0])
				}
				w.Header().Set("Cache-Control", "max-age=0, must-revalidate")
			}
			page := newPage("")
			page.Name = "home"
			page.Scheduler = s
			page.Write(w, req)
			return
		} else if req.URL.Path == "/hue/" {
			page := newPage("")
			page.Name = "hue"
			page.Scheduler = s
			page.Write(w, req)
			return
		} else if req.URL.Path == "/schedule/" {
			page := newPage("")
			page.Name = "schedule"
			page.Scheduler = s
			page.Write(w, req)
			return
		}
		StaticHandler(w, req)
	}
}

func StaticHandler(w http.ResponseWriter, req *http.Request) {
	var filename = path.Join(*StaticRoot, req.URL.Path)
	f, err := os.Open(filename)
	if err != nil {
		log.Print(err)
		NotFoundHandler(w, req)
		return
	} else {
		defer f.Close()
	}

	setCacheControl(w, req)

	http.ServeFile(w, req, filename)
}
