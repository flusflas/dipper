package godotted_test

import (
	"fmt"
	"godotted"
	"reflect"
	"testing"
	"time"
)

type Publication struct {
	ISBN string
}

type Author struct {
	Name      string
	BirthDate time.Time
}

type Book struct {
	Title  string
	Year   int
	Author Author
	Genres []string
	Extra  map[interface{}]interface{}
	Publication
}

func mustParseDate(date string) time.Time {
	t, _ := time.Parse("2006-01-02", date)
	return t
}

func TestGetAttribute(t *testing.T) {

	testStruct := &Book{
		Title: "El nombre de la rosa",
		Year:  1980,
		Author: Author{
			Name:      "Umberto Eco",
			BirthDate: mustParseDate("1932-07-05"),
		},
		Genres: []string{"Mystery", "Crime"},
		Extra: map[interface{}]interface{}{
			"foo": map[string]int{
				"bar": 123,
			},
		},
		Publication: Publication{
			ISBN: "1234567890",
		},
	}

	type args struct {
		v         interface{}
		attribute string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "struct",
			args: args{
				v:         testStruct,
				attribute: "Author",
			},
			want: Author{
				Name:      "Umberto Eco",
				BirthDate: mustParseDate("1932-07-05"),
			},
		},
		{
			name: "map 1",
			args: args{
				v:         testStruct,
				attribute: "Extra.foo",
			},
			want: map[string]int{"bar": 123},
		},
		{
			name: "map 2",
			args: args{
				v:         testStruct,
				attribute: "Extra.foo.bar",
			},
			want: 123,
		},
		{
			name: "map 3",
			args: args{
				v: map[interface{}]interface{}{
					"foo": map[string]int{
						"bar": 123,
					},
					"bar": map[int]string{
						1: "a",
						2: "b",
					},
				},
				attribute: "bar.1",
			},
			want: godotted.ErrMapKeyNotString,
		},
		{
			name: "slice",
			args: args{
				v:         testStruct,
				attribute: "Genres.1",
			},
			want: "Crime",
		},
		{
			name: "slice root",
			args: args{
				v: []interface{}{
					123,
					map[string]interface{}{
						"x": 1,
						"y": 2,
					},
				},
				attribute: "1.x",
			},
			want: 1,
		},
		{
			name: "not found",
			args: args{
				v:         testStruct,
				attribute: "foo",
			},
			want: godotted.ErrNotFound,
		},
		{
			name: "invalid index",
			args: args{
				v:         testStruct,
				attribute: "Genres.a",
			},
			want: godotted.ErrInvalidIndex,
		},
		{
			name: "index out of range",
			args: args{
				v:         testStruct,
				attribute: "Genres.2",
			},
			want: godotted.ErrIndexOutOfRange,
		},
		{
			name: "negative index",
			args: args{
				v:         testStruct,
				attribute: "Genres.-1",
			},
			want: godotted.ErrIndexOutOfRange,
		},
		{
			name: "unexported",
			args: args{
				v:         testStruct,
				attribute: "Author.BirthDate.wall",
			},
			want: godotted.ErrUnexported,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := godotted.GetAttribute(tt.args.v, tt.args.attribute)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAttributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAttributes(t *testing.T) {
	type args struct {
		v          interface{}
		attributes []string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "struct",
			args: args{
				v: Book{
					Title: "El nombre de la rosa",
					Year:  1980,
					Author: Author{
						Name:      "Umberto Eco",
						BirthDate: mustParseDate("1932-07-05"),
					},
					Genres: []string{"Mystery", "Crime"},
					Extra: map[interface{}]interface{}{
						"foo": map[string]int{
							"bar": 123,
						},
					},
					Publication: Publication{
						ISBN: "1234567890",
					},
				},
				attributes: []string{
					"Author",
					"Author.BirthDate",
					"Name", // does not exist
					"Publication.ISBN",
					"Genres.1",
					"Author.BirthDate.wall", // unexported field
					"Extra.foo",
				},
			},
			want: map[string]interface{}{
				"Author": Author{
					Name:      "Umberto Eco",
					BirthDate: mustParseDate("1932-07-05"),
				},
				"Author.BirthDate":      mustParseDate("1932-07-05"),
				"Name":                  godotted.ErrNotFound,
				"Publication.ISBN":      "1234567890",
				"Genres.1":              "Crime",
				"Author.BirthDate.wall": godotted.ErrUnexported,
				"Extra.foo":             map[string]int{"bar": 123},
			},
		},
		{
			name: "map",
			args: args{
				v: map[interface{}]interface{}{
					"foo": map[string]int{
						"bar": 123,
					},
					"bar": [][]interface{}{
						{"a", "b", "c"},
						{1000, 2000},
						{
							map[string]interface{}{
								"hello": "hola",
								"bye":   "adiós",
								"extra": []float64{-10.0, 27.3, 5.5, 100.0},
							},
							map[int]string{1: "abc", 2: "def"},
						},
					},
				},
				attributes: []string{
					"bar.0",
					"foo.bar",
					"bar.1.1",
					"bar.2.0.bye",
					"bar.2.0.extra.2",
					"bar.3",         // out of range
					"bar.2.1.1",     // not a map[string]
					"foo.bar.value", // does not exist
					"foo.x",         // does not exist
				},
			},
			want: map[string]interface{}{
				"bar.0":           []interface{}{"a", "b", "c"},
				"foo.bar":         123,
				"bar.1.1":         2000,
				"bar.2.0.bye":     "adiós",
				"bar.2.0.extra.2": 5.5,
				"bar.3":           godotted.ErrIndexOutOfRange,
				"bar.2.1.1":       godotted.ErrMapKeyNotString,
				"foo.bar.value":   godotted.ErrNotFound,
				"foo.x":           godotted.ErrNotFound,
			},
		},
		{
			name: "slice",
			args: args{
				v: []interface{}{
					"foo",
					map[string]int{
						"bar": 123,
					},
				},
				attributes: []string{
					"0",
					"1.bar",
					"1",
				},
			},
			want: map[string]interface{}{
				"0":     "foo",
				"1.bar": 123,
				"1":     map[string]int{"bar": 123},
			},
		},
		{
			name: "nil",
			args: args{
				v: nil,
				attributes: []string{
					"foo",
				},
			},
			want: map[string]interface{}{
				"foo": godotted.ErrNotFound,
			},
		},
		{
			name: "no attributes",
			args: args{
				v: map[string]interface{}{
					"foo": 123,
					"bar": 456,
				},
				attributes: []string{},
			},
			want: map[string]interface{}{},
		},
		{
			name: "pointer",
			args: args{
				v: &map[string]interface{}{
					"w": &Book{
						Title: "Rendezvous with Rama",
						Year:  1972,
					},
					"x": 123,
					"y": func() *string {
						s := "foobar"
						return &s
					}(),
					"z": func() *int {
						return nil
					}(),
				},
				attributes: []string{
					"w",
					"x",
					"y",
					"z",
					"y.a", // does not exist
					"z.a", // does not exist
				},
			},
			want: map[string]interface{}{
				"w": &Book{
					Title: "Rendezvous with Rama",
					Year:  1972,
				},
				"x": 123,
				"y": func() *string {
					s := "foobar"
					return &s
				}(),
				"z": func() *int {
					return nil
				}(),
				"y.a": godotted.ErrNotFound,
				"z.a": godotted.ErrNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := godotted.GetAttributes(tt.args.v, tt.args.attributes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAttributes() = %v, want %v", got, tt.want)
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
