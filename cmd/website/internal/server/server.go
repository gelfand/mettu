package server

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gelfand/log"
	"github.com/gelfand/mettu/cmd/website/internal/config"
	"github.com/gelfand/mettu/cmd/website/internal/dathtml"
	"github.com/gelfand/mettu/repo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	mux       *chi.Mux
	templates map[string]*template.Template
	db        *repo.DB
}

func NewServer(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		cfg = config.DefaultConfig
	}
	db, err := repo.NewDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	s := &Server{
		mux:       chi.NewMux(),
		templates: dathtml.LoadHTML(),
		db:        db,
	}

	return s, nil
}

func (s *Server) ListenAndServeTLS(addr, certFile, keyFile string) {
	http.ListenAndServeTLS(addr, certFile, keyFile, s.mux)
}

func (s *Server) Install() {
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.CleanPath)
	s.mux.Use(middleware.NoCache)
	// mux.Use(middleware.BasicAuth("Alpha Leak", creds))

	s.mux.Mount("/static/css/", http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))
	s.mux.Mount("/static/js/", http.StripPrefix("/static/js/", http.FileServer(http.Dir("static/js"))))
	s.mux.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		f, err := os.Open("./static/index.html")
		if err != nil {
			log.Fatalf("Unable to read index html template: %v", err)
		}

		buf := make([]byte, 8192)
		n, _ := f.Read(buf)
		w.Write(buf[:n])
	})
	s.mux.Route("/swaps", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		})
	})
	s.mux.HandleFunc("/exchanges", func(w http.ResponseWriter, r *http.Request) {
		tx, err := s.db.BeginRo(r.Context())
		if err != nil {
			log.Error("could not begin transaction: %v", err)
			return
		}
		defer tx.Rollback()

		exchs, err := s.db.AllExchanges(tx)
		if err != nil {
			log.Fatal(err)
		}
		s.templates["exchanges"].Execute(w, exchs)
	})
	s.mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		tx, err := s.db.BeginRo(r.Context())
		if err != nil {
			log.Errorf("could not begin database transaction: %v", err)
			return
		}
		defer tx.Rollback()

		accs, err := s.db.AllAccounts(tx)
		if err != nil {
			log.Fatal(err)
		}

		t := s.templates["accounts"]
		t.Execute(w, accs)
	})
}
