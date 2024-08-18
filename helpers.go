package main

import (
	"html/template"
	"net/http"
)

func (s *Server) notFound(
	w http.ResponseWriter,
	r *http.Request,
	title string,
	message string,
) {
	w.WriteHeader(http.StatusNotFound)
	s.renderTemplate(
		w, r,
		struct {
			Title   string
			Message string
		}{
			Title:   title,
			Message: message,
		},
		"layout",
		"html/layout.html",
		"html/notfound.html")
}

func (s *Server) badRequest(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message string,
) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func (s *Server) serverError(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Ops something went wrong. Please check the server logs."))
}

func (s *Server) renderTemplate(
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	name string,
	files ...string,
) {
	t, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderMessage(
	w http.ResponseWriter,
	r *http.Request,
	title string,
	paragraphs ...interface{},
) {
	s.renderTemplate(
		w, r,
		struct {
			Title      string
			Paragraphs []interface{}
		}{
			Title:      title,
			Paragraphs: paragraphs,
		},
		"layout",
		"html/layout.html",
		"html/message.html",
	)
}
