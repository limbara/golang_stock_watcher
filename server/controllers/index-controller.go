package controllers

import (
	"html/template"
	"net/http"
	"path"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	var filepath = path.Join("views", "index.html")
	var tmpl, err = template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		panic(err)
	}

	var data = map[string]interface{}{
		"title": "Stock Watcher",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
