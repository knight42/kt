package main

import "testing"

func TestNormalizeColor(t *testing.T) {
	cases := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"auto":            {input: "auto", want: "auto"},
		"always":          {input: "always", want: "always"},
		"never":           {input: "never", want: "never"},
		"on maps always":  {input: "on", want: "always"},
		"yes maps always": {input: "yes", want: "always"},
		"off maps never":  {input: "off", want: "never"},
		"no maps never":   {input: "no", want: "never"},
		"invalid":         {input: "bogus", wantErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := normalizeColor(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got none", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("normalizeColor(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
