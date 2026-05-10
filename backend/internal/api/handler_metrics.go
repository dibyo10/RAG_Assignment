package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dibyochakraborty/notebooklm/internal/store"
)

type metricsHandler struct {
	metricsStore *store.MetricsStore
}

func (h *metricsHandler) Global(c *gin.Context) {
	data, err := h.metricsStore.GetGlobalMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []*store.DayMetrics{}
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *metricsHandler) Session(c *gin.Context) {
	id := c.Param("id")
	logs, err := h.metricsStore.GetSessionMetrics(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if logs == nil {
		logs = []*store.QueryLog{}
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *metricsHandler) Document(c *gin.Context) {
	id := c.Param("id")
	data, err := h.metricsStore.GetDocumentMetrics(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if data == nil {
		data = []*store.DayMetrics{}
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
