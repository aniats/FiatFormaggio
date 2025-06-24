package app

import "testing"

func TestParseUserResponse(t *testing.T) {
	tests := []struct {
		input    string
		expected ResponseType
	}{
		{"да", ResponseYes},
		{"yes", ResponseYes},
		{"ага", ResponseYes},
		{"ок", ResponseYes},
		{"подтверждаю", ResponseYes},

		{"дв", ResponseYes},  // "да" with typo
		{"yea", ResponseYes}, // "yes" with typo
		{"окк", ResponseYes}, // "ок" with extra char

		{"нет", ResponseNo},
		{"no", ResponseNo},
		{"не", ResponseNo},
		{"отказываюсь", ResponseNo},

		{"нкт", ResponseNo},
		{"но", ResponseNo},

		{"отмена", ResponseCancel},
		{"cancel", ResponseCancel},
		{"стоп", ResponseCancel},
		{"назад", ResponseCancel},

		{"отменв", ResponseCancel},
		{"стпп", ResponseCancel},

		{"пропустить", ResponseSkip},
		{"skip", ResponseSkip},
		{"пропуск", ResponseSkip},
		{"дальше", ResponseSkip},

		{"пропуститъ", ResponseSkip},
		{"skup", ResponseSkip},

		{"привет", ResponseUnknown},
		{"test", ResponseUnknown},
		{"1234", ResponseUnknown},
		{"", ResponseUnknown},
		{"xyz", ResponseUnknown},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ParseUserResponse(test.input)
			if result != test.expected {
				t.Errorf("ParseUserResponse(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

func TestNormalizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  ДА  ", "да"},
		{"YES!", "yes"},
		{"нет.", "нет"},
		{"OK?", "ok"},
		{"  пропустить,  ", "пропустить"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := normalizeInput(test.input)
			if result != test.expected {
				t.Errorf("normalizeInput(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected int
	}{
		{"да", "дв", 1},
		{"нет", "нкт", 1},
		{"ок", "окк", 1},
		{"стоп", "стп", 1},
		{"да", "да", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"cat", "bat", 1},
		{"kitten", "sitting", 3},
	}

	for _, test := range tests {
		t.Run(test.s1+"_"+test.s2, func(t *testing.T) {
			result := LevenshteinDistance(test.s1, test.s2)
			if result != test.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, expected %d", test.s1, test.s2, result, test.expected)
			}
		})
	}
}

func TestBooleanHelpers(t *testing.T) {
	positiveInputs := []string{"да", "yes", "ок", "дв", "yea"}
	for _, input := range positiveInputs {
		if !IsPositiveResponse(input) {
			t.Errorf("IsPositiveResponse(%q) should return true", input)
		}
	}

	negativeInputs := []string{"нет", "no", "отмена", "нкт", "но"}
	for _, input := range negativeInputs {
		if !IsNegativeResponse(input) {
			t.Errorf("IsNegativeResponse(%q) should return true", input)
		}
	}

	skipInputs := []string{"пропустить", "skip", "дальше", "skup"}
	for _, input := range skipInputs {
		if !IsSkipResponse(input) {
			t.Errorf("IsSkipResponse(%q) should return true", input)
		}
	}

	validInputs := []string{"да", "нет", "пропустить", "отмена", "дв", "нкт"}
	for _, input := range validInputs {
		if !IsValidResponse(input) {
			t.Errorf("IsValidResponse(%q) should return true", input)
		}
	}

	invalidInputs := []string{"привет", "test", "xyz", "1234"}
	for _, input := range invalidInputs {
		if IsValidResponse(input) {
			t.Errorf("IsValidResponse(%q) should return false", input)
		}
	}
}
