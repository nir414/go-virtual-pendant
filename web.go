package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// webInterfaceHandler serves the web interface using templates
func webInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template loading error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// staticFileHandler serves static files (CSS, JS, images)
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	// URL에서 /static/ 제거
	file := r.URL.Path[len("/static/"):]
	filePath := filepath.Join("static", file)

	// 보안을 위해 경로 검증
	if filepath.Dir(filePath) != "static" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, filePath)
}
