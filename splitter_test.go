package dipper

import (
	"reflect"
	"testing"
)

func TestAttributeSplitter(t *testing.T) {
	type args struct {
		s   string
		sep string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "01",
			args: args{s: "foo.bar", sep: "."},
			want: []string{"foo", "bar"},
		},
		{
			name: "02",
			args: args{s: ",a,b,c,d,e,f,g,h,", sep: ","},
			want: []string{"", "a", "b", "c", "d", "e", "f", "g", "h", ""},
		},
		{
			name: "03",
			args: args{s: "", sep: "."},
			want: []string{""},
		},
		{
			name: "04",
			args: args{s: "-a->b->c->d->e->f->g->h->", sep: "->"},
			want: []string{"-a", "b", "c", "d", "e", "f", "g", "h", ""},
		},
		{
			name: "05",
			args: args{s: "", sep: "->"},
			want: []string{""},
		},
		{
			name: "06",
			args: args{s: "a-->b", sep: "-->"},
			want: []string{"a", "b"},
		},
		{
			name: "07",
			args: args{s: "", sep: "."},
			want: []string{""},
		},
		{
			name: "08",
			args: args{s: "Book", sep: "."},
			want: []string{"Book"},
		},
		{
			name: "09",
			args: args{s: "Book.1.Year", sep: "."},
			want: []string{"Book", "1", "Year"},
		},
		{
			name: "10",
			args: args{s: "Book[1].Year", sep: "."},
			want: []string{"Book", "[1]", "Year"},
		},
		{
			name: "11",
			args: args{s: "genres[id=0]", sep: "."},
			want: []string{"genres", "[id=0]"},
		},
		{
			name: "12",
			args: args{s: "genres[id=0].name", sep: "."},
			want: []string{"genres", "[id=0]", "name"},
		},
		{
			name: "13",
			args: args{s: "genres[id=0.0].name", sep: "."},
			want: []string{"genres", "[id=0.0]", "name"},
		},
		{
			name: "14",
			args: args{s: "genres->[id=0.0]->name", sep: "->"},
			want: []string{"genres", "", "[id=0.0]", "name"},
		},
		{
			name: "15",
			args: args{s: "genres.[id=0].name", sep: "."},
			want: []string{"genres", "", "[id=0]", "name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := newAttributeSplitter(tt.args.s, tt.args.sep)

			count := split.CountRemaining()
			if count != len(tt.want) {
				t.Errorf("CountRemaining() returned an unexpected value (got %d, want %d)", count, len(tt.want))
			}

			expectedIndex := 0
			var results []string
			for split.HasMore() {
				part, index := split.Next()
				results = append(results, part)
				if index != expectedIndex {
					t.Errorf("HasMore() returned an unexpected index (got %d, want %d)", index, expectedIndex)
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
