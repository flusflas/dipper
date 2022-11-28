package godotted_test

import (
	"encoding/json"
	"fmt"
	"godotted"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	type args struct {
		obj       interface{}
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
				obj:       getTestStruct(),
				attribute: "Author.Name",
			},
			want: "Umberto Eco",
		},
		{
			name: "map",
			args: args{
				obj:       getTestStruct(),
				attribute: "Extra.foo.bar",
			},
			want: 123,
		},
		{
			name: "map with dotted keys",
			args: args{
				obj: map[string]interface{}{
					"1.0": []string{"Initial release", "Buf fix"},
				},
				attribute: "1.0.0",
			},
			want: godotted.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := godotted.Get(tt.args.obj, tt.args.attribute)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMany(t *testing.T) {
	type args struct {
		obj        interface{}
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
				obj: getTestStruct(),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := godotted.GetMany(tt.args.obj, tt.args.attributes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMany() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {

	type args struct {
		obj       interface{}
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
				attribute: "Publication.ISBN",
				obj:       getTestStruct(),
				newValue:  "9788845207051",
			},
			want: want{
				result:   nil,
				newValue: "9788845207051",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := godotted.Set(tt.args.obj, tt.args.attribute, tt.args.newValue)
			if !reflect.DeepEqual(got, tt.want.result) {
				t.Errorf("Set() = %v, want %v", got, tt.want)
			}
			if tt.want.result == nil {
				newValue := godotted.Get(tt.args.obj, tt.args.attribute)
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
