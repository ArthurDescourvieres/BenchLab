package main

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, svc *SensorService) {
	h := NewHandler(svc)
	g := r.Group("/sensors")
	{
		g.POST("", h.Create)
		g.GET("", h.List)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
