package dipper_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/flusflas/dipper"
)

type Publication struct {
	ISBN string `json:"isbn"`
}

type Author struct {
	Name      string    `json:"name"`
	BirthDate time.Time `json:"birth_date"`
}

type Genre struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Book struct {
	Title      string                 `json:"title"`
	Year       int                    `json:"year"`
	Author     Author                 `json:"author"`
	GenreNames []string               `json:"genre_names"`
	Genres     []Genre                `json:"genres"`
	Extra      map[string]interface{} `json:"extra"`
	Any        interface{}            `json:"any"`
	Publication
}

func intPtr(v int) *int { return &v }

func toJSONMap(v interface{}) map[string]interface{} {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		panic(err)
	}
	return m
}

func mustParseDate(date string) time.Time {
	t, _ := time.Parse("2006-01-02", date)
	return t
}

func getTestStruct() *Book {
	return &Book{
		Title: "El nombre de la rosa",
		Year:  1980,
		Author: Author{
			Name:      "Umberto Eco",
			BirthDate: mustParseDate("1932-07-05"),
		},
		GenreNames: []string{"Mystery", "Crime"},
		Genres: []Genre{
			{
				ID:          0,
				Name:        "Mystery",
				Description: "Fiction genre where the nature of an event, usually a murder or other crime, remains mysterious until the end of the story",
			},
			{
				ID:          1,
				Name:        "Crime",
				Description: "Narratives that centre on criminal acts and especially on the investigation of a crime, often a murder",
			},
		},
		Extra: map[string]interface{}{
			"foo": map[string]int{
				"bar": 123,
			},
		},
		Publication: Publication{
			ISBN: "1234567890",
		},
	}
}

func TestDipper_Get(t *testing.T) {
	type args struct {
		obj       interface{}
		attribute string
	}
	tests := []struct {
		name      string
		separator string // Default is "."
		args      args
		want      interface{}
	}{
		{
			name: "empty attribute",
			args: args{
				obj:       getTestStruct(),
				attribute: "",
			},
			want: getTestStruct(),
		},
		{
			name: "struct",
			args: args{
				obj:       getTestStruct(),
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
				obj:       getTestStruct(),
				attribute: "Extra.foo",
			},
			want: map[string]int{"bar": 123},
		},
		{
			name:      "map 2",
			separator: "/",
			args: args{
				obj:       getTestStruct(),
				attribute: "Extra/foo/bar",
			},
			want: 123,
		},
		{
			name: "map 3",
			args: args{
				obj: map[string]interface{}{
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
			want: dipper.ErrMapKeyNotString,
		},
		{
			name:      "map with dotted keys",
			separator: "/",
			args: args{
				obj: map[string]interface{}{
					"1.0": []string{"Initial release", "Buf fix"},
				},
				attribute: "1.0/0",
			},
			want: "Initial release",
		},
		{
			name:      "map with empty key attributes",
			separator: ".",
			args: args{
				obj: map[string]interface{}{
					"": []int{1, 2, 3},
				},
				attribute: "[2]",
			},
			want: 3,
		},
		{
			name:      "map with empty key attributes (nested)",
			separator: ".",
			args: args{
				obj: map[string]interface{}{
					"": map[string]interface{}{
						"": []int{1, 2, 3},
					},
				},
				attribute: ".[2]",
			},
			want: 3,
		},
		{
			name: "get struct from slice",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres.1",
			},
			want: Genre{
				ID:          1,
				Name:        "Crime",
				Description: "Narratives that centre on criminal acts and especially on the investigation of a crime, often a murder",
			},
		},
		{
			name: "get field of struct from slice",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres.1.Name",
			},
			want: "Crime",
		},
		{
			name: "slice",
			args: args{
				obj: []interface{}{
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
			name: "slice using brackets notation from root",
			args: args{
				obj:       []int{1, 2, 3},
				attribute: "[1]",
			},
			want: 2,
		},
		{
			name: "slice using brackets notation",
			args: args{
				obj:       getTestStruct(),
				attribute: "GenreNames[1]",
			},
			want: "Crime",
		},
		{
			name: "slice of pointers should return a pointer",
			args: args{
				obj: []*Book{
					getTestStruct(),
				},
				attribute: "[0]",
			},
			want: getTestStruct(),
		},
		{
			name: "slice using brackets notation after separator",
			args: args{
				obj:       getTestStruct(),
				attribute: "GenreNames.[1]",
			},
			want: dipper.ErrInvalidIndex,
		},
		{
			name: "not found",
			args: args{
				obj:       getTestStruct(),
				attribute: "foo",
			},
			want: dipper.ErrNotFound,
		},
		{
			name: "invalid index",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres.a",
			},
			want: dipper.ErrInvalidIndex,
		},
		{
			name: "index out of range",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres.2",
			},
			want: dipper.ErrIndexOutOfRange,
		},
		{
			name: "negative index",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres.-1",
			},
			want: dipper.ErrIndexOutOfRange,
		},
		{
			name: "unexported",
			args: args{
				obj:       *getTestStruct(),
				attribute: "Author.BirthDate.wall",
			},
			want: dipper.ErrUnexported,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dipper.New(dipper.Options{Separator: tt.separator})
			got := d.Get(tt.args.obj, tt.args.attribute)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDipper_GetMany(t *testing.T) {
	type args struct {
		obj        interface{}
		attributes []string
	}
	tests := []struct {
		name      string
		separator string // Default is "."
		args      args
		want      dipper.Fields
	}{
		{
			name:      "struct",
			separator: "->",
			args: args{
				obj: getTestStruct(),
				attributes: []string{
					"Author",
					"Author->BirthDate",
					"Name", // does not exist
					"Publication->ISBN",
					"GenreNames->1",
					"Author->BirthDate->wall", // unexported field
					"Extra->foo",
				},
			},
			want: map[string]interface{}{
				"Author": Author{
					Name:      "Umberto Eco",
					BirthDate: mustParseDate("1932-07-05"),
				},
				"Author->BirthDate":       mustParseDate("1932-07-05"),
				"Name":                    dipper.ErrNotFound,
				"Publication->ISBN":       "1234567890",
				"GenreNames->1":           "Crime",
				"Author->BirthDate->wall": dipper.ErrUnexported,
				"Extra->foo":              map[string]int{"bar": 123},
			},
		},
		{
			name: "map",
			args: args{
				obj: map[string]interface{}{
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
				"bar.3":           dipper.ErrIndexOutOfRange,
				"bar.2.1.1":       dipper.ErrMapKeyNotString,
				"foo.bar.value":   dipper.ErrNotFound,
				"foo.x":           dipper.ErrNotFound,
			},
		},
		{
			name: "slice",
			args: args{
				obj: []interface{}{
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
				obj: nil,
				attributes: []string{
					"foo",
				},
			},
			want: map[string]interface{}{
				"foo": dipper.ErrNotFound,
			},
		},
		{
			name: "no attributes",
			args: args{
				obj: map[string]interface{}{
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
				obj: &map[string]interface{}{
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
				"y.a": dipper.ErrNotFound,
				"z.a": dipper.ErrNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dipper.New(dipper.Options{Separator: tt.separator})
			got := d.GetMany(tt.args.obj, tt.args.attributes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMany() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDipper_Set(t *testing.T) {
	type args struct {
		v         interface{}
		attribute string
		newValue  interface{}
	}
	type want struct {
		result   interface{}
		deleted  bool
		newValue interface{}
	}
	tests := []struct {
		name      string
		separator string // Default is "."
		args      args
		want      want
	}{
		{
			name: "update int value in struct",
			args: args{
				attribute: "Year",
				v: &Book{
					Title: "El nombre de la rosa",
					Year:  1979,
				},
				newValue: 1980,
			},
			want: want{
				result:   nil,
				newValue: 1980,
			},
		},
		{
			name: "update map value",
			args: args{
				attribute: "1",
				v: map[string]interface{}{
					"1": "Rendezvous with Rama",
				},
				newValue: Book{
					Title: "El nombre de la rosa",
				},
			},
			want: want{
				result: nil,
				newValue: Book{
					Title: "El nombre de la rosa",
				},
			},
		},
		{
			name: "update nested map value",
			args: args{
				attribute: "Extra.1",
				v: &Book{
					Extra: map[string]interface{}{
						"1": "Rendezvous with Rama",
					},
				},
				newValue: Book{
					Title: "El nombre de la rosa",
				},
			},
			want: want{
				result: nil,
				newValue: Book{
					Title: "El nombre de la rosa",
				},
			},
		},
		{
			name: "update interface value",
			args: args{
				attribute: "Any",
				v: &Book{
					Any: "some value",
				},
				newValue: "different value",
			},
			want: want{
				result:   nil,
				newValue: "different value",
			},
		},
		{
			name: "update slice element",
			args: args{
				attribute: "1",
				v:         []interface{}{"1", 2, 3.0},
				newValue:  2 + 0i,
			},
			want: want{
				result:   nil,
				newValue: 2 + 0i,
			},
		},
		{
			name: "update addressable array element",
			args: args{
				attribute: "1",
				v:         &[3]interface{}{"1", 2, 3.0},
				newValue:  2 + 0i,
			},
			want: want{
				result:   nil,
				newValue: 2 + 0i,
			},
		},
		{
			name: "addressable int value",
			args: args{
				attribute: "",
				v:         intPtr(1979),
				newValue:  intPtr(1980),
			},
			want: want{
				result:   nil,
				newValue: intPtr(1980),
			},
		},
		{
			name:      "update map with dotted keys",
			separator: "/",
			args: args{
				attribute: "1.0/1.beta",
				v: map[string]interface{}{
					"1.0.0.beta": "Initial release",
					"1.0": map[string]interface{}{
						"1.beta": "Buf fix",
						"2.beta": "Another buf fix",
					},
				},
				newValue: "First bug fix",
			},
			want: want{
				result:   nil,
				newValue: "First bug fix",
			},
		},
		{
			name: "delete map key",
			args: args{
				attribute: "foo",
				v: map[string]int{
					"foo": 123,
				},
				newValue: dipper.Delete,
			},
			want: want{
				result:   nil,
				deleted:  true,
				newValue: dipper.Delete,
			},
		},
		{
			name: "delete slice element",
			args: args{
				attribute: "1",
				v:         []int{1, 2, 3},
				newValue:  dipper.Delete,
			},
			want: want{
				result:   nil,
				newValue: 0,
			},
		},
		{
			name: "set zero value to []interface",
			args: args{
				attribute: "3",
				v:         []interface{}{"1", 2, 3.0, 4 + 0i},
				newValue:  dipper.Zero,
			},
			want: want{
				result:   nil,
				newValue: nil,
			},
		},
		{
			name: "set zero value to string in struct",
			args: args{
				attribute: "Title",
				v: &Book{
					Title: "El nombre de la rosa",
				},
				newValue: dipper.Zero,
			},
			want: want{
				result:   nil,
				newValue: "",
			},
		},
		{
			name:      "set key to nil map",
			separator: "/",
			args: args{
				attribute: "Extra/foo",
				v:         &Book{Extra: nil},
				newValue:  123,
			},
			want: want{
				result:   nil,
				newValue: 123,
			},
		},
		{
			name: "set key to nil slice",
			args: args{
				attribute: "Genres.0",
				v:         &Book{Genres: nil},
				newValue:  "Sci-Fi",
			},
			want: want{
				result: dipper.ErrIndexOutOfRange,
			},
		},
		{
			name: "field not match",
			args: args{
				attribute: "Name",
				v: &Book{
					Title: "El nombre de la rosa",
					Year:  1979,
				},
				newValue: 1980,
			},
			want: want{
				result:   dipper.ErrNotFound,
				newValue: 1980,
			},
		},
		{
			name: "unaddressable struct value",
			args: args{
				attribute: "Year",
				v: Book{
					Title: "El nombre de la rosa",
					Year:  1979,
				},
				newValue: 1980,
			},
			want: want{
				result:   dipper.ErrUnaddressable,
				newValue: 1980,
			},
		},
		{
			name: "unaddressable int value",
			args: args{
				attribute: "",
				v:         1979,
				newValue:  1980,
			},
			want: want{
				result:   dipper.ErrUnaddressable,
				newValue: 1980,
			},
		},
		{
			name: "unaddressable array",
			args: args{
				attribute: "1",
				v:         [3]interface{}{"1", 2, 3.0},
				newValue:  2 + 0i,
			},
			want: want{
				result:   dipper.ErrUnaddressable,
				newValue: 2 + 0i,
			},
		},
		{
			name: "update struct in map with wrong type",
			args: args{
				attribute: "1",
				v: map[string]string{
					"1": "Rendezvous with Rama",
				},
				newValue: 123,
			},
			want: want{
				result:   dipper.ErrTypesDoNotMatch,
				newValue: 123,
			},
		},
		{
			name: "types do not match",
			args: args{
				attribute: "Year",
				v: &Book{
					Title: "El nombre de la rosa",
					Year:  1979,
				},
				newValue: "1980",
			},
			want: want{
				result:   dipper.ErrTypesDoNotMatch,
				newValue: "1980",
			},
		},
		{
			name: "update map value with invalid key type",
			args: args{
				attribute: "1",
				v: map[int]interface{}{
					1: "Rendezvous with Rama",
				},
				newValue: "El nombre de la rosa",
			},
			want: want{
				result: dipper.ErrMapKeyNotString,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dipper.New(dipper.Options{Separator: tt.separator})
			got := d.Set(tt.args.v, tt.args.attribute, tt.args.newValue)
			if !reflect.DeepEqual(got, tt.want.result) {
				t.Errorf("Set() = %v, want %v", got, tt.want)
			}
			if tt.want.result == nil {
				newValue := d.Get(tt.args.v, tt.args.attribute)
				if tt.want.deleted && newValue != dipper.ErrNotFound {
					t.Errorf("Set() => Map value was not deleted")
				}

				if !tt.want.deleted && !reflect.DeepEqual(newValue, tt.want.newValue) {
					t.Errorf("Set() => Value did not change to %v", tt.want.newValue)
				}
			}
		})
	}
}
