package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

var site = template.Must(template.ParseFiles("templates/site.html"))

func makeTemplate(names ...string) *template.Template {
	t, err := site.Clone()
	if err != nil {
		log.Fatal("cloning site: ", err)
	}
	return template.Must(t.ParseFiles(names...)) // TODO: allow overriding?
}

type View interface {
	Prefix() string
	Name() string
	Match(req *http.Request) bool
	Data(req *http.Request) Data
}

type Data map[string]interface{}

type view struct {
	prefix, name string
	data         Data
}

func (v *view) Prefix() string {
	return v.prefix
}

func (v *view) Name() string {
	return v.name
}

func (v *view) Match(req *http.Request) bool {
	return req.URL.Path == v.Prefix()
}

func (v *view) Data(req *http.Request) Data {
	if v.data == nil {
		v.data = make(Data)
	}
	v.data["Title"] = v.Name()
	return v.data
}

func add(view View) {
	t := makeTemplate("templates/" + view.Name() + ".html")
	http.HandleFunc(view.Prefix(), func(w http.ResponseWriter, req *http.Request) {
		var d Data
		if view.Match(req) {
			d = view.Data(req)
		} else {
			w.Header().Set("Cache-Control", "max-age=10, must-revalidate")
			w.WriteHeader(http.StatusNotFound)
		}
		var bw bytes.Buffer
		h := md5.New()
		mw := io.MultiWriter(&bw, h)
		err := t.ExecuteTemplate(mw, "html", d)
		if err == nil {
			w.Header().Set("ETag", fmt.Sprintf(`"%x"`, h.Sum(nil)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", bw.Len()))
			w.Write(bw.Bytes())
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func ListenAndServe(address string, scheduler *scheduler) {
	add(&view{prefix: "/", name: "home", data: Data{"Scheduler": scheduler}})
	add(&view{prefix: "/hue/", name: "hue", data: Data{"Scheduler": scheduler}})
	add(&view{prefix: "/schedule/", name: "schedule", data: Data{"Scheduler": scheduler}})
	add(&view{prefix: "/transition/", name: "transition", data: Data{"Scheduler": scheduler}})
	http.HandleFunc("/bootstrap/", StaticHandler)
	http.HandleFunc("/jquery/", StaticHandler)
	http.HandleFunc("/post", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			if err := req.ParseForm(); err == nil {
				name := req.Form["do_transition"]
				scheduler.Hue.Do(name[0])
			}
			// TODO: write a response
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Print("ListenAndServe:", err)
	}
}
