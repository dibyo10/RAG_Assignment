package embedder

import (
	"context"
	"fmt"
	"sync"

	"github.com/dibyochakraborty/notebooklm/internal/chunker"
)

type EmbeddedChunk struct {
	Chunk     chunker.ChunkResult
	Embedding []float32
	Index     int // original position in input slice
}

type embedJob struct {
	index int
	chunk chunker.ChunkResult
}

type embedResult struct {
	index     int
	embedding []float32
	err       error
}

// EmbedChunks embeds all chunks concurrently using a worker pool.
// numWorkers controls how many goroutines call the OpenAI API simultaneously.
func (e *Embedder) EmbedChunks(ctx context.Context, chunks []chunker.ChunkResult, numWorkers int) ([]EmbeddedChunk, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	if numWorkers <= 0 {
		numWorkers = 8
	}
	if numWorkers > len(chunks) {
		numWorkers = len(chunks)
	}

	jobsChan := make(chan embedJob, len(chunks))
	resultsChan := make(chan embedResult, len(chunks))

	// Fan-out: spawn workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsChan {
				select {
				case <-ctx.Done():
					resultsChan <- embedResult{index: job.index, err: ctx.Err()}
					return
				default:
				}
				vecs, err := e.Embed(ctx, []string{job.chunk.Text})
				if err != nil {
					resultsChan <- embedResult{index: job.index, err: fmt.Errorf("chunk %d: %w", job.index, err)}
					return
				}
				resultsChan <- embedResult{index: job.index, embedding: vecs[0]}
			}
		}()
	}

	// Dispatch jobs
	for i, c := range chunks {
		jobsChan <- embedJob{index: i, chunk: c}
	}
	close(jobsChan)

	// Close results chan when all workers done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Fan-in: collect results in order
	ordered := make([]EmbeddedChunk, len(chunks))
	for res := range resultsChan {
		if res.err != nil {
			return nil, res.err
		}
		ordered[res.index] = EmbeddedChunk{
			Chunk:     chunks[res.index],
			Embedding: res.embedding,
			Index:     res.index,
		}
	}
	return ordered, nil
}
