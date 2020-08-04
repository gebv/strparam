package httprouter

import (
	"context"
	"reflect"
	"testing"
)

func TestParsedParamsFromCtx(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want map[string]string
	}{
		{"", context.Background(), map[string]string{}},
		{"", setParsedParamsToCtx(context.Background(), map[string]string{"foo": "bar"}), map[string]string{"foo": "bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParsedParamsFromCtx(tt.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsedParamsFromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}
