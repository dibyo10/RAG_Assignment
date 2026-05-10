package chunker

import "strings"

type ChunkResult struct {
	Text       string
	StartChar  int
	EndChar    int
	ChunkIndex int
}

var defaultSeparators = []string{"\n\n", "\n", ". ", "? ", "! ", " ", ""}

// Chunk splits text using recursive character splitting with overlap.
// chunkSize is the target max chars per chunk; overlap is chars carried from the previous chunk.
func Chunk(text string, chunkSize, overlap int) []ChunkResult {
	pieces := splitRecursive(text, chunkSize, defaultSeparators)
	merged := mergeWithOverlap(pieces, chunkSize, overlap)

	// Compute character positions in the original text
	results := make([]ChunkResult, 0, len(merged))
	searchFrom := 0
	for i, m := range merged {
		// Strip overlap prefix for position tracking (find the non-overlapping core)
		start := strings.Index(text[searchFrom:], m)
		if start == -1 {
			start = 0
		} else {
			start += searchFrom
		}
		end := start + len(m)
		if end > len(text) {
			end = len(text)
		}
		results = append(results, ChunkResult{
			Text:       m,
			StartChar:  start,
			EndChar:    end,
			ChunkIndex: i,
		})
		// Advance past the non-overlapping portion
		if end > searchFrom {
			searchFrom = end - overlap
			if searchFrom < 0 {
				searchFrom = 0
			}
		}
	}
	return results
}

func splitRecursive(text string, chunkSize int, separators []string) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}
	if len(separators) == 0 {
		// Hard split at chunkSize
		var parts []string
		for len(text) > 0 {
			end := chunkSize
			if end > len(text) {
				end = len(text)
			}
			parts = append(parts, text[:end])
			text = text[end:]
		}
		return parts
	}

	sep := separators[0]
	rest := separators[1:]

	if sep == "" {
		return splitRecursive(text, chunkSize, rest)
	}

	splits := strings.Split(text, sep)
	var result []string
	for _, s := range splits {
		if s == "" {
			continue
		}
		if len(s) <= chunkSize {
			result = append(result, s)
		} else {
			result = append(result, splitRecursive(s, chunkSize, rest)...)
		}
	}
	return result
}

func mergeWithOverlap(pieces []string, chunkSize, overlap int) []string {
	var chunks []string
	current := ""

	for _, p := range pieces {
		candidate := current
		if candidate != "" {
			candidate += " "
		}
		candidate += p

		if len(candidate) > chunkSize && current != "" {
			chunks = append(chunks, current)
			// Start new chunk with overlap from end of current
			if overlap > 0 && len(current) > overlap {
				current = current[len(current)-overlap:] + " " + p
			} else {
				current = p
			}
		} else {
			current = candidate
		}
	}
	if current != "" {
		chunks = append(chunks, current)
	}
	return chunks
}
