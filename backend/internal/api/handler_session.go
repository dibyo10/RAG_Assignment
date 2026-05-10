package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dibyochakraborty/notebooklm/internal/store"
)

type sessionHandler struct {
	sessStore *store.SessionStore
	docStore  *store.DocumentStore
}

func (h *sessionHandler) Create(c *gin.Context) {
	var body struct {
		DocumentID string `json:"document_id" binding:"required"`
		Title      string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.docStore.Get(body.DocumentID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	sess, err := h.sessStore.Create(body.DocumentID, body.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": sess})
}

func (h *sessionHandler) List(c *gin.Context) {
	docID := c.Query("document_id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_id query param required"})
		return
	}
	sessions, err := h.sessStore.ListByDocument(docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sessions == nil {
		sessions = []*store.Session{}
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *sessionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	sess, err := h.sessStore.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	msgs, err := h.sessStore.GetMessages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if msgs == nil {
		msgs = []*store.Message{}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"session": sess, "messages": msgs}})
}

func (h *sessionHandler) UpdateTitle(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.sessStore.UpdateTitle(id, body.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"updated": true}})
}

func (h *sessionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.sessStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}
