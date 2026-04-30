package query

import (
	"bytes"
	"testing"
)

func TestParse_Match(t *testing.T) {
	tests := map[string]struct {
		query   string
		line    string
		want    bool
	}{
		"single keyword match": {
			query: "error",
			line:  "an error occurred",
			want:  true,
		},
		"single keyword no match": {
			query: "error",
			line:  "all good",
			want:  false,
		},
		"case insensitive": {
			query: "Error",
			line:  "an ERROR occurred",
			want:  true,
		},
		"and both match": {
			query: "error and fatal",
			line:  "fatal error occurred",
			want:  true,
		},
		"and one missing": {
			query: "error and fatal",
			line:  "an error occurred",
			want:  false,
		},
		"or first matches": {
			query: "error or warning",
			line:  "an error occurred",
			want:  true,
		},
		"or second matches": {
			query: "error or warning",
			line:  "a warning issued",
			want:  true,
		},
		"or neither matches": {
			query: "error or warning",
			line:  "all good",
			want:  false,
		},
		"and has higher precedence than or": {
			query: "a or b and c",
			line:  "a",
			want:  true,
		},
		"and has higher precedence than or - needs both for and": {
			query: "a or b and c",
			line:  "b",
			want:  false,
		},
		"parentheses override precedence": {
			query: "(a or b) and c",
			line:  "b c",
			want:  true,
		},
		"parentheses override precedence - missing c": {
			query: "(a or b) and c",
			line:  "a b",
			want:  false,
		},
		"quoted keyword with space": {
			query: `"error code" and fatal`,
			line:  "fatal error code 500",
			want:  true,
		},
		"quoted keyword no match": {
			query: `"error code"`,
			line:  "error happened, code 500",
			want:  false,
		},
		"nested parens": {
			query: "(a or b) and (c or d)",
			line:  "b d",
			want:  true,
		},
		"quoted and is a keyword not operator": {
			query: `"and" or error`,
			line:  "and",
			want:  true,
		},
		"quoted or is a keyword not operator": {
			query: `"or"`,
			line:  "or happened",
			want:  true,
		},
		"quoted or does not act as operator": {
			query: `"or"`,
			line:  "something else",
			want:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			expr, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.query, err)
			}
			got := expr.Match([]byte(tt.line))
			if got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := map[string]string{
		"empty":                "",
		"unterminated quote":   `"hello`,
		"missing closing paren": "(a or b",
		"unexpected token":     "and",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", input)
			}
		})
	}
}

func TestTerms(t *testing.T) {
	tests := map[string]struct {
		query string
		want  []string
	}{
		"single keyword": {
			query: "error",
			want:  []string{"error"},
		},
		"and expression": {
			query: "error and fatal",
			want:  []string{"error", "fatal"},
		},
		"or expression": {
			query: "error or warning",
			want:  []string{"error", "warning"},
		},
		"nested": {
			query: "(a or b) and c",
			want:  []string{"a", "b", "c"},
		},
		"quoted keyword": {
			query: `"error code" and fatal`,
			want:  []string{"error code", "fatal"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			expr, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.query, err)
			}
			got := expr.Terms()
			if len(got) != len(tt.want) {
				t.Fatalf("Terms() returned %d terms, want %d", len(got), len(tt.want))
			}
			for i, want := range tt.want {
				if !bytes.Equal(got[i], []byte(want)) {
					t.Errorf("Terms()[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestHighlight(t *testing.T) {
	hl := func(s string) string {
		return "\033[1;31m" + s + "\033[0m"
	}

	tests := map[string]struct {
		line  string
		terms []string
		want  string
	}{
		"no terms": {
			line:  "hello world",
			terms: nil,
			want:  "hello world",
		},
		"single match": {
			line:  "an error occurred",
			terms: []string{"error"},
			want:  "an " + hl("error") + " occurred",
		},
		"case insensitive": {
			line:  "an ERROR occurred",
			terms: []string{"error"},
			want:  "an " + hl("ERROR") + " occurred",
		},
		"multiple terms": {
			line:  "fatal error occurred",
			terms: []string{"fatal", "error"},
			want:  hl("fatal") + " " + hl("error") + " occurred",
		},
		"no match": {
			line:  "all good",
			terms: []string{"error"},
			want:  "all good",
		},
		"overlapping prefers longer": {
			line:  "error_code found",
			terms: []string{"err", "error_code"},
			want:  hl("error_code") + " found",
		},
		"multiple occurrences": {
			line:  "error and error again",
			terms: []string{"error"},
			want:  hl("error") + " and " + hl("error") + " again",
		},
		"match at start": {
			line:  "error!",
			terms: []string{"error"},
			want:  hl("error") + "!",
		},
		"match at end": {
			line:  "got error",
			terms: []string{"error"},
			want:  "got " + hl("error"),
		},
		"quoted multi-word term": {
			line:  "fatal error code 500",
			terms: []string{"error code"},
			want:  "fatal " + hl("error code") + " 500",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var terms [][]byte
			for _, s := range tt.terms {
				terms = append(terms, []byte(s))
			}
			got := Highlight([]byte(tt.line), terms)
			if !bytes.Equal(got, []byte(tt.want)) {
				t.Errorf("Highlight() = %q, want %q", got, tt.want)
			}
		})
	}
}
