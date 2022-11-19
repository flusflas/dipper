package godotted_test

import (
	"fmt"
	"godotted"
	"testing"
)

func TestFieldError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    error
		want string
	}{
		{
			name: "not found",
			e:    godotted.ErrNotFound,
			want: "godotted: field not found",
		},
		{
			name: "unexported",
			e:    godotted.ErrUnexported,
			want: "godotted: field is unexported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func TestError(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want error
	}{
		{
			name: "nil interface",
			arg:  nil,
			want: nil,
		},
		{
			name: "no error value",
			arg:  123,
			want: nil,
		},
		{
			name: "non-field error",
			arg:  fmt.Errorf("some error"),
			want: nil,
		},
		{
			name: "field error",
			arg:  godotted.ErrNotFound,
			want: godotted.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := godotted.Error(tt.arg); got != tt.want {
				t.Errorf("IsFieldError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFields_HasErrors(t *testing.T) {
	tests := []struct {
		name string
		f    godotted.Fields
		want bool
	}{
		{
			name: "no errors",
			f: map[string]interface{}{
				"x": 1,
				"y": 2,
			},
			want: false,
		},
		{
			name: "one field error",
			f: map[string]interface{}{
				"x":   1,
				"y.5": godotted.ErrIndexOutOfRange,
			},
			want: true,
		},
		{
			name: "one non-field error",
			f: map[string]interface{}{
				"x": 1,
				"y": fmt.Errorf("some error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFields_FirstError(t *testing.T) {
	tests := []struct {
		name string
		f    godotted.Fields
		want error
	}{
		{
			name: "no errors",
			f: map[string]interface{}{
				"x": 1,
				"y": 2,
			},
			want: nil,
		},
		{
			name: "one field error",
			f: map[string]interface{}{
				"x":   1,
				"y.5": godotted.ErrIndexOutOfRange,
			},
			want: godotted.ErrIndexOutOfRange,
		},
		{
			name: "one non-field error",
			f: map[string]interface{}{
				"x": 1,
				"y": fmt.Errorf("some error"),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f.FirstError()
			if err != tt.want {
				t.Errorf("FirstError() error = %v, want %v", err, tt.want)
			}
		})
	}
}
