package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/worker"
)

type Server struct {
	Addr        string
	db          *storage.Storage
	worker      *worker.Worker
	cache       map[string]interface{}
	cache_mutex *sync.Mutex

	BasePath string

	// auth
	Username string
	Password string
	// https
	CertFile string
	KeyFile  string
	public   bool
}

func NewServer(db *storage.Storage, addr string, public bool) *Server {
	var w *worker.Worker
	if !public {
		w = worker.NewWorker(db)
	}

	return &Server{
		db:          db,
		Addr:        addr,
		worker:      w,
		cache:       make(map[string]interface{}),
		cache_mutex: &sync.Mutex{},
		public:      public,
	}
}

func (h *Server) GetAddr() string {
	proto := "http"
	if h.CertFile != "" && h.KeyFile != "" {
		proto = "https"
	}
	return proto + "://" + h.Addr + h.BasePath
}

func (s *Server) Start() {
	if !s.public {
		refreshRate := s.db.GetSettingsValueInt64("refresh_rate")
		s.worker.FindFavicons()
		s.worker.StartFeedCleaner()
		s.worker.SetRefreshRate(refreshRate)
		if refreshRate > 0 {
			s.worker.RefreshFeeds()
		}
	} else {
		log.Println("running in public mode")
	}

	httpserver := &http.Server{Addr: s.Addr, Handler: s.handler()}

	var err error
	if s.CertFile != "" && s.KeyFile != "" {
		err = httpserver.ListenAndServeTLS(s.CertFile, s.KeyFile)
	} else {
		err = httpserver.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
