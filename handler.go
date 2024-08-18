package main

import (
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"strings"
	"time"
)

func (s *Server) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method == "GET" || r.Method == "HEAD" {
		s.handleGET(w, r)
		return
	}
	if r.Method == "POST" && r.URL.Path == "/" {
		s.handlePOST(w, r)
		return
	}
}

func (s *Server) handlePOST(
	w http.ResponseWriter,
	r *http.Request,
) {
	mediaType := r.Header.Get("Content-Type")
	if mediaType != "application/x-www-form-urlencoded" {
		s.badRequest(
			w, r,
			http.StatusUnsupportedMediaType,
			"Invalid media type posted.")
		return
	}

	err := r.ParseForm()
	if err != nil {
		s.badRequest(
			w, r,
			http.StatusBadRequest,
			"Invalid form data posted.")
		return
	}
	form := r.PostForm

	message := form.Get("message")
	destruct := true
	ttl := time.Hour * 24 * 365

	note := &Note{
		Data:     []byte(message),
		Destruct: destruct,
	}

	key := uuid.NewString()
	err = s.RedisCache.Set(
		&cache.Item{
			Ctx:            r.Context(),
			Key:            key,
			Value:          note,
			TTL:            ttl,
			SkipLocalCache: true,
		})
	if err != nil {
		fmt.Println(err)
		s.serverError(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	noteURL := fmt.Sprintf("%s/%s", s.BaseURL, key)

	data := struct {
		NoteURL string
	}{
		NoteURL: noteURL,
	}

	s.renderTemplate(w, r, data, "layout", "html/layout.html", "html/success.html")
}

func (s *Server) handleGET(
	w http.ResponseWriter,
	r *http.Request,
) {
	path := r.URL.Path
	if path == "/" {
		s.renderTemplate(
			w, r, nil,
			"layout",
			"html/layout.html",
			"html/index.html")
		return
	}

	noteID := strings.TrimPrefix(path, "/")

	ctx := r.Context()
	note := &Note{}
	err := s.RedisCache.GetSkippingLocalCache(
		ctx,
		noteID,
		note)
	if err != nil {
		s.notFound(
			w, r,
			"Note Not Found",
			fmt.Sprintf("Note with ID %s does not exist.", noteID))
		return
	}

	data := template.HTML(string(note.Data))

	if note.Destruct {
		err := s.RedisCache.Delete(ctx, noteID)
		if err != nil {
			fmt.Println(err)
			s.serverError(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

	s.renderTemplate(
		w, r, struct {
			Title       string
			NoteContent template.HTML
		}{
			Title:       "Note",
			NoteContent: data,
		},
		"layout",
		"html/layout.html",
		"html/note.html")
	return
}
