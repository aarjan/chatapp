package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type templateHandler struct {
	templ    *template.Template
	filename string
	once     sync.Once
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	if err := t.templ.Execute(w, r); err != nil {
		log.Fatal("Error rendering template: ", err)
	}
}

func main() {
	addr := flag.String("addr", ":8080", "port address")
	flag.Parse()

	r := newRoom()
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting server on address", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
