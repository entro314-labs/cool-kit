package api

import (
	"encoding/json"
	"testing"
)

func TestID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    ID
		wantErr bool
	}{
		{"String ID", `"123"`, ID("123"), false},
		{"Int ID", `456`, ID("456"), false},
		{"Float ID", `789.0`, ID("789"), false},
		{"Empty String", `""`, ID(""), false},
		{"Invalid", `{}`, ID(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id ID
			err := json.Unmarshal([]byte(tt.json), &id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id != tt.want {
				t.Errorf("ID.UnmarshalJSON() = %v, want %v", id, tt.want)
			}
		})
	}
}

func TestParseLogs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty",
			input: "",
			want:  "",
		},
		{
			name:  "Single Array",
			input: `[{"output":"line1\n","hidden":false},{"output":"line2\n","hidden":false}]`,
			want:  "line1\n\nline2\n",
		},
		{
			name:  "Concatenated Arrays",
			input: `[{"output":"start\n","hidden":false}][{"output":"end\n","hidden":false}]`,
			want:  "start\n\nend\n",
		},
		{
			name:  "Mixed Objects",
			input: `{"output":"single\n","hidden":false}`,
			want:  "single\n",
		},
		{
			name:  "Malformed JSON fallback",
			input: `Raw string output`,
			want:  "Raw string output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLogs(tt.input)
			if got != tt.want {
				t.Errorf("ParseLogs() = %q, want %q", got, tt.want)
			}
		})
	}
}
