package godotted

import (
	"reflect"
	"testing"
)

func TestAttributeSplitter(t *testing.T) {
	type args struct {
		s   string
		sep byte
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "01",
			args: args{s: "foo.bar", sep: '.'},
			want: []string{"foo", "bar"},
		},
		{
			name: "02",
			args: args{s: ".a.b.c.d.e.f.g.h.", sep: '.'},
			want: []string{"", "a", "b", "c", "d", "e", "f", "g", "h", ""},
		},
		{
			name: "03",
			args: args{s: "", sep: '.'},
			want: []string{""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := newAttributeSplitter(tt.args.s, tt.args.sep)

			expectedIndex := 0
			var results []string
			for split.HasMore() {
				part, index := split.Next()
				results = append(results, part)
				if index != expectedIndex {
					t.Errorf("HasMore() returned an expected index (got %d, want %d)", index, expectedIndex)
				}
				expectedIndex++
			}

			if _, i := split.Next(); i != -1 {
				t.Errorf("Next() returned an index different than -1, but splitter is exhausted")
			}

			if !reflect.DeepEqual(results, tt.want) {
				t.Errorf("newAttributeSplitter() = %v, want %v", results, tt.want)
			}
		})
	}
}
