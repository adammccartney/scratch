package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed templates/*.html
var files embed.FS

var funcMap = template.FuncMap{
	"hasPrefix": strings.HasPrefix,
}

var tmpl = template.Must(template.New("").Funcs(funcMap).ParseFS(files, "templates/*.html"))

type IndexHandler struct{}

func (i IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	core := "This is a network test"
	data := []string{}
	j := 1
	for j <= 750 {
		data = append(data, core)
		j = j + 1
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func mylog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Handler function called - %T\n", h)
		h.ServeHTTP(w, r)
	})
}

func main() {

	srv := http.Server{
		Addr: "0.0.0.0:9868",
	}
	index := IndexHandler{}
	http.Handle("/", mylog(index))
	srv.ListenAndServe()

}
