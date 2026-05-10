package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dibyochakraborty/notebooklm/internal/ingestion"
	"github.com/dibyochakraborty/notebooklm/internal/parser"
	"github.com/dibyochakraborty/notebooklm/internal/store"
)

type documentHandler struct {
	docStore  *store.DocumentStore
	pipeline  *ingestion.Pipeline
	uploadDir string
}

func (h *documentHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field required"})
		return
	}
	defer file.Close()

	mimeType := parser.DetectMIME(header.Filename)
	if mimeType == "application/octet-stream" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only PDF and text files are supported"})
		return
	}

	// Save to disk
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload dir"})
		return
	}
	ext := filepath.Ext(header.Filename)
	filePath := filepath.Join(h.uploadDir, uuid.NewString()+ext)
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	if _, err := dst.ReadFrom(file); err != nil {
		dst.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}
	dst.Close()

	// Create document record
	doc, err := h.docStore.Create(header.Filename, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("db error: %v", err)})
		return
	}

	// Background indexing — detached from request context
	go h.pipeline.Run(context.Background(), doc.ID, filePath, mimeType)

	c.JSON(http.StatusAccepted, gin.H{"data": doc})
}

func (h *documentHandler) List(c *gin.Context) {
	docs, err := h.docStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if docs == nil {
		docs = []*store.Document{}
	}
	c.JSON(http.StatusOK, gin.H{"data": docs})
}

func (h *documentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	doc, err := h.docStore.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	count, _ := h.docStore.CountChunks(id)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"document": doc, "chunk_count": count}})
}

func (h *documentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.docStore.Get(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if err := h.docStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}
