package metrics

import (
	"math"

	"github.com/dibyochakraborty/notebooklm/internal/retriever"
)

type Result struct {
	MRR       float64
	RecallAtK float64
	NDCG      float64
	ScoreMin  float64
	ScoreMax  float64
	ScoreMean float64
	ScoreStd  float64
	Count     int
}

// Compute calculates retrieval metrics using pseudo-relevance labeling.
// The top chunk is always pseudo-relevant; additionally any chunk with
// score > mean + 0.5*std is also marked relevant.
func Compute(chunks []*retriever.ScoredChunk) Result {
	if len(chunks) == 0 {
		return Result{}
	}

	scores := make([]float64, len(chunks))
	for i, c := range chunks {
		scores[i] = c.Score
	}

	mean, std := meanStd(scores)
	threshold := mean + 0.5*std

	// Pseudo-relevance: mark chunks above threshold as relevant; always include rank-1
	relevant := make([]bool, len(chunks))
	relevant[0] = true
	relevantCount := 1
	for i := 1; i < len(chunks); i++ {
		if scores[i] > threshold {
			relevant[i] = true
			relevantCount++
		}
	}

	// MRR: reciprocal rank of first relevant item (always rank 1)
	mrr := 1.0

	// Recall@K
	recallatK := float64(relevantCount) / float64(relevantCount) // always 1.0 if top-1 is relevant
	// More meaningful: how many relevant items we retrieved vs total possible
	// Since all relevant are in our retrieved set (we just defined them from it),
	// we use a threshold-based denominator across the full set
	if relevantCount > 0 {
		recallatK = float64(relevantCount) / float64(len(chunks))
		if recallatK > 1 {
			recallatK = 1
		}
	}

	// NDCG using binary relevance
	dcg := 0.0
	idcg := 0.0
	// Ideal: place all relevant items at top ranks
	idealRel := make([]float64, len(chunks))
	for i := 0; i < relevantCount && i < len(idealRel); i++ {
		idealRel[i] = 1.0
	}
	for i, rel := range relevant {
		gain := 0.0
		if rel {
			gain = 1.0
		}
		dcg += gain / math.Log2(float64(i+2))
		idcg += idealRel[i] / math.Log2(float64(i+2))
	}
	ndcg := 0.0
	if idcg > 0 {
		ndcg = dcg / idcg
	}

	minScore, maxScore := scores[0], scores[0]
	for _, s := range scores {
		if s < minScore {
			minScore = s
		}
		if s > maxScore {
			maxScore = s
		}
	}

	return Result{
		MRR:       mrr,
		RecallAtK: recallatK,
		NDCG:      ndcg,
		ScoreMin:  minScore,
		ScoreMax:  maxScore,
		ScoreMean: mean,
		ScoreStd:  std,
		Count:     len(chunks),
	}
}

func meanStd(vals []float64) (float64, float64) {
	if len(vals) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))
	if len(vals) == 1 {
		return mean, 0
	}
	variance := 0.0
	for _, v := range vals {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(vals) - 1)
	return mean, math.Sqrt(variance)
}
