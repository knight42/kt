package query

import "testing"

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
