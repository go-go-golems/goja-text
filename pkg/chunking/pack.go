package chunking

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func Pack(spans []Span, options PackOptions) (*PackResult, error) {
	options = defaultPackOptions(options)
	if err := validatePackOptions(options); err != nil {
		return nil, err
	}
	weights := make([]int, len(spans))
	for i, span := range spans {
		weight, err := measureText(span.Text, options.Measure)
		if err != nil {
			return nil, err
		}
		weights[i] = weight
	}
	return packWithWeights(spans, weights, options.MaxUnits, options.Overlap.Value, options.Oversized,
		strategy("pack", map[string]any{
			"maxUnits":  options.MaxUnits,
			"measure":   options.Measure,
			"overlap":   map[string]any{"unit": "spans", "value": options.Overlap.Value},
			"oversized": options.Oversized,
		}))
}

func PackWeighted(items []WeightedSpan, options WeightedPackOptions) (*PackResult, error) {
	if options.MaxWeight <= 0 {
		return nil, fmt.Errorf("chunking.packWeighted: maxWeight must be greater than zero")
	}
	if options.OverlapWeight < 0 {
		return nil, fmt.Errorf("chunking.packWeighted: overlapWeight must be nonnegative")
	}
	if options.Oversized == "" {
		options.Oversized = "allow"
	}
	if options.Oversized != "allow" && options.Oversized != "error" {
		return nil, fmt.Errorf("chunking.packWeighted: oversized must be allow or error")
	}
	spans := make([]Span, len(items))
	weights := make([]int, len(items))
	for i, item := range items {
		if item.Weight < 0 {
			return nil, fmt.Errorf("chunking: invalid_weight: item %d has negative weight", i)
		}
		spans[i] = item.Span
		weights[i] = item.Weight
	}
	return packWeightedBudget(spans, weights, options)
}

func packWeightedBudget(spans []Span, weights []int, options WeightedPackOptions) (*PackResult, error) {
	result, err := packWithWeights(spans, weights, options.MaxWeight, 0, options.Oversized,
		strategy("packWeighted", map[string]any{
			"maxWeight":     options.MaxWeight,
			"overlapWeight": options.OverlapWeight,
			"oversized":     options.Oversized,
		}))
	if err != nil || options.OverlapWeight == 0 || len(result.Chunks) < 2 {
		return result, err
	}
	// Weighted overlap is reconstructed from the prior chunk's trailing complete spans.
	for i := 1; i < len(result.Chunks); i++ {
		prior := result.Chunks[i-1]
		firstOrdinal := result.Chunks[i].SpanOrdinals[0]
		overlap := make([]int, 0)
		overlapWeight := 0
		for j := len(prior.SpanOrdinals) - 1; j >= 0; j-- {
			ordinal := prior.SpanOrdinals[j]
			if ordinal >= firstOrdinal || weights[ordinal] > options.OverlapWeight-overlapWeight {
				continue
			}
			overlap = append([]int{ordinal}, overlap...)
			overlapWeight += weights[ordinal]
		}
		for len(overlap) > 0 && overlapWeight+result.Chunks[i].Weight > options.MaxWeight {
			overlapWeight -= weights[overlap[0]]
			overlap = overlap[1:]
		}
		if len(overlap) == 0 {
			continue
		}
		ordinals := append(overlap, result.Chunks[i].SpanOrdinals...)
		result.Chunks[i] = makeChunk(spans, weights, ordinals, i)
	}
	return result, nil
}

func packWithWeights(spans []Span, weights []int, budget, overlap int, oversized string, spec StrategySpec) (*PackResult, error) {
	if len(spans) != len(weights) {
		return nil, fmt.Errorf("chunking: invalid_weight: spans and weights differ in length")
	}
	result := &PackResult{Spec: spec, Chunks: []PackedChunk{}, Diagnostics: []Diagnostic{}}
	current := make([]int, 0)
	currentWeight := 0
	emit := func() {
		if len(current) == 0 {
			return
		}
		result.Chunks = append(result.Chunks, makeChunk(spans, weights, current, len(result.Chunks)))
	}
	for i, weight := range weights {
		if weight < 0 {
			return nil, fmt.Errorf("chunking: invalid_weight: span %d has negative weight", i)
		}
		if len(current) == 0 && weight > budget {
			if oversized == "error" {
				return nil, fmt.Errorf("chunking: span_exceeds_budget: span %d weighs %d, budget %d", i, weight, budget)
			}
			current = []int{i}
			emit()
			result.Chunks[len(result.Chunks)-1].Oversized = true
			diagnostic := Diagnostic{Code: "span_exceeds_budget", Severity: "warning", Message: fmt.Sprintf("span %d exceeds budget", i), StartByte: spans[i].StartByte, EndByte: spans[i].EndByte}
			result.Chunks[len(result.Chunks)-1].Diagnostics = []Diagnostic{diagnostic}
			result.Diagnostics = append(result.Diagnostics, diagnostic)
			current = current[:0]
			currentWeight = 0
			continue
		}
		if len(current) == 0 || currentWeight+weight <= budget {
			current = append(current, i)
			currentWeight += weight
			continue
		}
		emit()
		keep := overlap
		if keep > len(current) {
			keep = len(current)
		}
		current = append([]int(nil), current[len(current)-keep:]...)
		currentWeight = sumWeights(weights, current)
		for len(current) > 0 && currentWeight+weight > budget {
			currentWeight -= weights[current[0]]
			current = current[1:]
		}
		current = append(current, i)
		currentWeight += weight
	}
	emit()
	return result, nil
}

func makeChunk(spans []Span, weights []int, ordinals []int, ordinal int) PackedChunk {
	var text strings.Builder
	chunk := PackedChunk{Ordinal: ordinal, SpanOrdinals: append([]int(nil), ordinals...), Diagnostics: []Diagnostic{}}
	for _, spanOrdinal := range ordinals {
		span := spans[spanOrdinal]
		text.WriteString(span.Text)
		chunk.Weight += weights[spanOrdinal]
		if len(chunk.HeadingPath) == 0 && len(span.HeadingPath) > 0 {
			chunk.HeadingPath = append([]string(nil), span.HeadingPath...)
		}
		if chunk.Level == "" && span.Level != "" {
			chunk.Level = span.Level
		}
	}
	first := spans[ordinals[0]]
	last := spans[ordinals[len(ordinals)-1]]
	chunk.Text = text.String()
	chunk.StartByte, chunk.EndByte = first.StartByte, last.EndByte
	chunk.StartRune, chunk.EndRune = first.StartRune, last.EndRune
	return chunk
}

func defaultPackOptions(options PackOptions) PackOptions {
	if options.Measure == "" {
		options.Measure = "runes"
	}
	if options.Overlap.Unit == "" {
		options.Overlap.Unit = "spans"
	}
	if options.Oversized == "" {
		options.Oversized = "allow"
	}
	return options
}

func validatePackOptions(options PackOptions) error {
	if options.MaxUnits <= 0 {
		return fmt.Errorf("chunking.pack: maxUnits must be greater than zero")
	}
	if options.Overlap.Unit != "spans" || options.Overlap.Value < 0 {
		return fmt.Errorf("chunking.pack: overlap must use nonnegative span units")
	}
	if options.Oversized != "allow" && options.Oversized != "error" {
		return fmt.Errorf("chunking.pack: oversized must be allow or error")
	}
	_, err := measureText("", options.Measure)
	return err
}

func measureText(text, measure string) (int, error) {
	switch measure {
	case "bytes":
		return len(text), nil
	case "runes":
		return utf8.RuneCountInString(text), nil
	case "words":
		return len(strings.Fields(text)), nil
	default:
		return 0, fmt.Errorf("chunking: unknown_measure: %q", measure)
	}
}

func sumWeights(weights []int, ordinals []int) int {
	total := 0
	for _, ordinal := range ordinals {
		total += weights[ordinal]
	}
	return total
}
