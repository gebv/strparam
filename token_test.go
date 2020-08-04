package strparam

import (
	"testing"
)

func TestToken_Equal(t *testing.T) {
	tests := []struct {
		name   string
		token1 Token
		token2 Token
		want   bool
	}{
		{"", ConstToken("C"), ConstToken("C"), true},
		{"", ConstToken("C"), ConstToken("c"), false},
		{"", ConstToken("C"), SeparatorToken("C"), false},
		{"", ConstToken("C"), SeparatorToken("c"), false},
		{"", ParameterToken("C"), ParsedParameterToken("C", "val"), false},
		{"", ParameterToken("C"), ParsedParameterToken("C", ""), false},
		{"", StartToken, EndToken, false},
		{"", EndToken, EndToken, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token1.Equal(tt.token2); got != tt.want {
				t.Errorf("Token.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_ParamName(t *testing.T) {
	tests := []struct {
		name  string
		token Token
		want  string
	}{
		{"", ConstToken("C"), ""},
		{"", SeparatorToken("C"), ""},
		{"", ParameterToken("C"), "C"},
		{"", ParameterToken("c"), "c"},
		{"", ParsedParameterToken("c", "val"), "c"},
		{"", ParsedParameterToken("C", "val"), "C"},
		{"", EndToken, ""},
		{"", StartToken, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.ParamName(); got != tt.want {
				t.Errorf("Token.ParamName() = %v, want %v", got, tt.want)
			}
		})
	}
}
