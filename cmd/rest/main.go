package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ArthurDescourvieres/BenchLab/internal/rest"
	"github.com/ArthurDescourvieres/BenchLab/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		if p[0] == ':' {
			addr = p
		} else {
			addr = ":" + p
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	s := store.NewMemoryStore()
	rest.RegisterRoutes(r, s)

	log.Printf("REST listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
