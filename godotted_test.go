package godotted_test

import (
	"encoding/json"
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
	Any    interface{}
	Publication
}

func intPtr(v int) *int { return &v }

func mustParseDate(date string) time.Time {
	t, _ := time.Parse("2006-01-02", date)
	return t
}

func TestGet(t *testing.T) {

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
			name: "empty attribute",
			args: args{
				v:         testStruct,
				attribute: "",
			},
			want: testStruct,
		},
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
			name: "map with dotted keys",
			args: args{
				v: map[string]interface{}{
					"1.0": []string{"Initial release", "Buf fix"},
				},
				attribute: "1.0.0",
			},
			want: "Initial release",
		},
		{
			name: "slice in struct",
			args: args{
				v:         testStruct,
				attribute: "Genres.1",
			},
			want: "Crime",
		},
		{
			name: "slice",
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
			got := godotted.Get(tt.args.v, tt.args.attribute)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMany(t *testing.T) {
	type args struct {
		v          interface{}
		attributes []string
	}
	tests := []struct {
		name string
		args args
		want godotted.Fields
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
			got := godotted.GetMany(tt.args.v, tt.args.attributes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMany() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {

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
		name string
		args args
		want want
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
					Extra: map[interface{}]interface{}{
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
			name: "delete map key",
			args: args{
				attribute: "foo",
				v: map[string]int{
					"foo": 123,
				},
				newValue: godotted.Delete,
			},
			want: want{
				result:   nil,
				deleted:  true,
				newValue: godotted.Delete,
			},
		},
		{
			name: "delete slice element",
			args: args{
				attribute: "1",
				v:         []int{1, 2, 3},
				newValue:  godotted.Delete,
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
				newValue:  godotted.Zero,
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
				newValue: godotted.Zero,
			},
			want: want{
				result:   nil,
				newValue: "",
			},
		},
		{
			name: "set key to nil map",
			args: args{
				attribute: "Extra.foo",
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
				result: godotted.ErrIndexOutOfRange,
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
				result:   godotted.ErrNotFound,
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
				result:   godotted.ErrUnaddressable,
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
				result:   godotted.ErrUnaddressable,
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
				result:   godotted.ErrUnaddressable,
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
				result:   godotted.ErrTypesDoNotMatch,
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
				result:   godotted.ErrTypesDoNotMatch,
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
				result: godotted.ErrMapKeyNotString,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := godotted.Set(tt.args.v, tt.args.attribute, tt.args.newValue)
			if !reflect.DeepEqual(got, tt.want.result) {
				t.Errorf("Set() = %v, want %v", got, tt.want)
			}
			if tt.want.result == nil {
				newValue := godotted.Get(tt.args.v, tt.args.attribute)
				if tt.want.deleted && newValue != godotted.ErrNotFound {
					t.Errorf("Set() => Map value was not deleted")
				}

				if !tt.want.deleted && !reflect.DeepEqual(newValue, tt.want.newValue) {
					t.Errorf("Set() => Value did not change to %v", tt.want.newValue)
				}
			}
		})
	}
}

func ExampleGet() {
	persons := []struct {
		Name  string
		Age   int
		About map[string]interface{}
	}{
		{
			Name: "Leela",
			Age:  25,
			About: map[string]interface{}{
				"spaceship pilot":  "Also can drive cars",
				"depth_perception": false,
			},
		},
		{
			Name: "Fry",
			Age:  1025,
			About: map[string]interface{}{
				"delivery":    "3/10",
				"game_player": 8,
				"powers":      []string{"Psychic immunity", "Caffeine"},
			},
		},
	}

	fmt.Println(godotted.Get(persons, "0.Name"))
	fmt.Println(godotted.Get(persons, "0.About.spaceship pilot"))
	fmt.Println(godotted.Get(persons, "1.Age"))
	fmt.Println(godotted.Get(persons, "1.About.powers.0"))
	fmt.Println(godotted.Get(persons, "1.Height"))
	fmt.Println(godotted.Get(persons, "2"))

	// Output:
	// Leela
	// Also can drive cars
	// 1025
	// Psychic immunity
	// godotted: field not found
	// godotted: index out of range
}

func ExampleGetMany() {
	persons := []struct {
		Name  string
		Age   int
		About map[string]interface{}
	}{
		{
			Name: "Leela",
			Age:  25,
			About: map[string]interface{}{
				"spaceship pilot":  "Also can drive cars",
				"depth_perception": false,
			},
		},
		{
			Name: "Fry",
			Age:  1025,
			About: map[string]interface{}{
				"delivery":    "3/10",
				"game_player": 8,
				"powers":      []string{"Psychic immunity", "Caffeine"},
			},
		},
	}

	fields := godotted.GetMany(persons, []string{
		"0.Name",
		"1.About.powers.0",
		"1.Height",
	})

	b, _ := json.MarshalIndent(fields, "", "  ")
	fmt.Println(string(b))

	// Output:
	// {
	//   "0.Name": "Leela",
	//   "1.About.powers.0": "Psychic immunity",
	//   "1.Height": "godotted: field not found"
	// }
}

func ExampleSet() {
	person := struct {
		Name  string
		Age   int
		About map[string]interface{}
	}{
		Name:  "Leela",
		Age:   25,
		About: map[string]interface{}{},
	}

	fmt.Println(godotted.Set(&person, "Name", "Amy"))
	fmt.Println(godotted.Set(&person, "Age", 21))
	fmt.Println(godotted.Set(&person, "About.rich", true))
	fmt.Println(godotted.Set(person, "", true))
	fmt.Println(person)

	// Output:
	// <nil>
	// <nil>
	// <nil>
	// godotted: field is unaddressable
	// {Amy 21 map[rich:true]}
}
