package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"usrsvc/internal/config"
	"usrsvc/internal/pkg/db"
	"usrsvc/internal/pkg/log"
	"usrsvc/internal/repository"
	th "usrsvc/internal/transport/http"
	"usrsvc/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	pool, err := db.NewPool(cfg.PGDSN)
	if err != nil {
		log.Error.Fatalf("db: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPgUserRepo(pool)
	uc := usecase.NewUserUC(repo)
	h := th.NewHandler(uc)
	r := th.NewRouter(h, cfg.CORSAllow)

	addr := ":" + cfg.Port
	log.Info.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error.Println(err)
		os.Exit(1)
	}
}
