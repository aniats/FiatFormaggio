package app

import (
	"strings"
	"unicode/utf8"

	"github.com/samber/lo"
)

type ResponseType int

const (
	ResponseUnknown ResponseType = iota
	ResponseYes
	ResponseNo
	ResponseCancel
	ResponseSkip
)

var (
	positiveResponses = []string{
		"да", "yes", "y", "ага", "угу", "конечно", "ok", "ок", "хорошо",
		"подтверждаю", "подтверждать", "создать", "создай", "давай", "1",
	}

	negativeResponses = []string{
		"нет", "no", "n", "не", "неа", "отказываюсь", "не хочу", "0",
	}

	cancelResponses = []string{
		"отмена", "cancel", "отменить", "стоп", "stop", "выход", "exit",
		"назад", "back", "отменяю",
	}

	skipResponses = []string{
		"пропустить", "skip", "пропуск", "дальше", "next", "без этого",
		"не нужно", "не надо", "пусто", "empty",
	}
)

func LevenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return utf8.RuneCountInString(s2)
	}
	if len(s2) == 0 {
		return utf8.RuneCountInString(s1)
	}

	runes1 := []rune(s1)
	runes2 := []rune(s2)

	len1, len2 := len(runes1), len(runes2)

	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if runes1[i-1] != runes2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len1][len2]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func normalizeInput(input string) string {
	normalized := strings.ToLower(strings.TrimSpace(input))

	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, ",", "")
	normalized = strings.ReplaceAll(normalized, "!", "")
	normalized = strings.ReplaceAll(normalized, "?", "")
	return normalized
}

func findClosestMatch(input string, candidates []string, maxDistance int) (string, bool) {
	input = normalizeInput(input)

	for _, candidate := range candidates {
		if normalizeInput(candidate) == input {
			return candidate, true // Exact match
		}
	}

	for _, candidate := range candidates {
		distance := LevenshteinDistance(input, normalizeInput(candidate))
		if distance <= maxDistance && distance > 0 {
			return candidate, true
		}
	}

	return "", false
}

func ParseUserResponse(input string) ResponseType {
	input = normalizeInput(input)

	if input == "" {
		return ResponseUnknown
	}

	if lo.Contains(positiveResponses, input) {
		return ResponseYes
	}
	if lo.Contains(negativeResponses, input) {
		return ResponseNo
	}
	if lo.Contains(cancelResponses, input) {
		return ResponseCancel
	}
	if lo.Contains(skipResponses, input) {
		return ResponseSkip
	}

	if _, found := findClosestMatch(input, positiveResponses, 1); found {
		return ResponseYes
	}
	if _, found := findClosestMatch(input, negativeResponses, 1); found {
		return ResponseNo
	}
	if _, found := findClosestMatch(input, cancelResponses, 1); found {
		return ResponseCancel
	}
	if _, found := findClosestMatch(input, skipResponses, 1); found {
		return ResponseSkip
	}

	return ResponseUnknown
}

func IsPositiveResponse(input string) bool {
	return ParseUserResponse(input) == ResponseYes
}

func IsNegativeResponse(input string) bool {
	response := ParseUserResponse(input)
	return response == ResponseNo || response == ResponseCancel
}

func IsSkipResponse(input string) bool {
	return ParseUserResponse(input) == ResponseSkip
}

func IsValidResponse(input string) bool {
	return ParseUserResponse(input) != ResponseUnknown
}

func GetSuggestionMessage() string {
	return `❓ Пожалуйста, используйте один из вариантов:
✅ Для подтверждения: "да", "yes", "ок", "подтверждаю"
❌ Для отказа: "нет", "no", "отмена"
⏭️ Для пропуска: "пропустить", "skip"`
}
