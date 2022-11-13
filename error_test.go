package godotted_test

import (
	"fmt"
	"godotted"
	"testing"
)

func TestIsFieldError(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want bool
	}{
		{
			name: "nil interface",
			arg:  nil,
			want: false,
		},
		{
			name: "no error value",
			arg:  123,
			want: false,
		},
		{
			name: "non-field error",
			arg:  fmt.Errorf("some error"),
			want: false,
		},
		{
			name: "field error",
			arg:  godotted.ErrNotFound,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := godotted.IsFieldError(tt.arg); got != tt.want {
				t.Errorf("IsFieldError() = %v, want %v", got, tt.want)
			}
		})
	}
}
