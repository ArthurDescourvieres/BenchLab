package rest

import (
	"github.com/ArthurDescourvieres/BenchLab/internal/store"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes enregistre les routes REST /sensors sur le moteur Gin.
func RegisterRoutes(r *gin.Engine, s store.Store) {
	h := NewHandler(s)
	g := r.Group("/sensors")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
