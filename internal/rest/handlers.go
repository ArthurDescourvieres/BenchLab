package rest

import (
	"errors"
	"net/http"

	"github.com/ArthurDescourvieres/BenchLab/internal/store"
	"github.com/gin-gonic/gin"
)

// Handler expose le CRUD Sensor sur HTTP (JSON).
type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

// sensorInput corps attendu pour création / mise à jour (champs métier).
type sensorInput struct {
	ID            string  `json:"id,omitempty"`
	Name          string  `json:"name" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	Location      string  `json:"location" binding:"required"`
	Unit          string  `json:"unit" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	LastValue     float64 `json:"last_value"`
	LastReadingAt string  `json:"last_reading_at" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	var in sensorInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := &store.Sensor{
		Name:          in.Name,
		Type:          in.Type,
		Location:      in.Location,
		Unit:          in.Unit,
		Status:        in.Status,
		LastValue:     in.LastValue,
		LastReadingAt: in.LastReadingAt,
	}
	ctx := c.Request.Context()
	if err := h.store.Create(ctx, s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := h.store.Get(ctx, s.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	s, err := h.store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": store.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.store.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []store.Sensor{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var in sensorInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.ID != "" && in.ID != id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id in body must match path or be omitted"})
		return
	}
	s := store.Sensor{
		ID:            id,
		Name:          in.Name,
		Type:          in.Type,
		Location:      in.Location,
		Unit:          in.Unit,
		Status:        in.Status,
		LastValue:     in.LastValue,
		LastReadingAt: in.LastReadingAt,
	}
	ctx := c.Request.Context()
	if err := h.store.Update(ctx, s); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": store.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.store.Get(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.store.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": store.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
