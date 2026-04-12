package main

import (
	"errors"
	"net/http"

	"github.com/ArthurDescourvieres/BenchLab/internal/sensorsvc"
	"github.com/ArthurDescourvieres/BenchLab/store"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *sensorsvc.SensorService
}

func NewHandler(svc *sensorsvc.SensorService) *Handler {
	return &Handler{svc: svc}
}

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

func (h *Handler) Create(ginCtx *gin.Context) {
	var body sensorInput
	if err := ginCtx.ShouldBindJSON(&body); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reqCtx := ginCtx.Request.Context()
	created, err := h.svc.Create(reqCtx, sensorsvc.SensorPayload{
		Name:          body.Name,
		Type:          body.Type,
		Location:      body.Location,
		Unit:          body.Unit,
		Status:        body.Status,
		LastValue:     body.LastValue,
		LastReadingAt: body.LastReadingAt,
	})
	if err != nil {
		writeSensorError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusCreated, created)
}

func (h *Handler) Get(ginCtx *gin.Context) {
	sensorID := ginCtx.Param("id")
	reqCtx := ginCtx.Request.Context()
	found, err := h.svc.Get(reqCtx, sensorID)
	if err != nil {
		writeSensorError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, found)
}

func (h *Handler) List(ginCtx *gin.Context) {
	sensors, err := h.svc.List(ginCtx.Request.Context())
	if err != nil {
		writeSensorError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, sensors)
}

func (h *Handler) Update(ginCtx *gin.Context) {
	sensorID := ginCtx.Param("id")
	var body sensorInput
	if err := ginCtx.ShouldBindJSON(&body); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reqCtx := ginCtx.Request.Context()
	updated, err := h.svc.Update(reqCtx, sensorID, sensorsvc.SensorPayload{
		ID:            body.ID,
		Name:          body.Name,
		Type:          body.Type,
		Location:      body.Location,
		Unit:          body.Unit,
		Status:        body.Status,
		LastValue:     body.LastValue,
		LastReadingAt: body.LastReadingAt,
	})
	if err != nil {
		writeSensorError(ginCtx, err)
		return
	}
	ginCtx.JSON(http.StatusOK, updated)
}

func (h *Handler) Delete(ginCtx *gin.Context) {
	sensorID := ginCtx.Param("id")
	err := h.svc.Delete(ginCtx.Request.Context(), sensorID)
	if err != nil {
		writeSensorError(ginCtx, err)
		return
	}
	ginCtx.Status(http.StatusNoContent)
}

func writeSensorError(ginCtx *gin.Context, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		ginCtx.JSON(http.StatusNotFound, gin.H{"error": store.ErrNotFound.Error()})
	case errors.Is(err, sensorsvc.ErrInvalidSensorType),
		errors.Is(err, sensorsvc.ErrInvalidStatus),
		errors.Is(err, sensorsvc.ErrInvalidReadingAt),
		errors.Is(err, sensorsvc.ErrInvalidName),
		errors.Is(err, sensorsvc.ErrIDMismatch),
		errors.Is(err, sensorsvc.ErrInvalidID):
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
