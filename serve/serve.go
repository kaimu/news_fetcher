package serve

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kaimu/news_fetcher/fetch"
)

type Reply struct {
	Term   string           `json:"term"`
	Result []fetch.NewsItem `json:"results"`
}

// Listen запускает web-сервер для поиска новостей в результатах агрегации
func Listen(addr, staticDirPath string, db *sqlx.DB) {
	server := &http.Server{Addr: addr}

	http.Handle("/", http.FileServer(http.Dir(staticDirPath)))
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		reply := Reply{}
		reply.Term = strings.TrimSpace(r.FormValue("term"))

		searchResults, err := search(reply.Term, db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		reply.Result = searchResults
		jsonWriter := json.NewEncoder(w)
		err = jsonWriter.Encode(reply)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	go func() {
		log.Printf("News HTTP-server is listening at: %v\n", addr)
		log.Fatalf("Stopped listening with an error: %v\n", server.ListenAndServe())
	}()

	// Подписываемся на сигнал SIGINT
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Received SIGINT, shutting down gracefully...")

	// Завершаем работу сервера
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	e := server.Shutdown(ctx)
	if e != nil {
		log.Printf("Gracefull-shutdown error: %v\n", e)
	} else {
		log.Println("News HTTP-server has shutted down")
	}
}

func search(substring string, db *sqlx.DB) (result []fetch.NewsItem, err error) {
	err = db.Select(&result, `SELECT * FROM news WHERE title LIKE $1 ORDER BY rowid DESC`, "%"+substring+"%")
	return
}
