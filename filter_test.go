package dipper_test

import (
	"github.com/flusflas/dipper"
	"reflect"
	"testing"
)

func TestDipper_GetWithFilter(t *testing.T) {
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
			name: "get primitive value from slice with filter",
			args: args{
				obj:       getTestStruct(),
				attribute: "GenreNames[='Crime']",
			},
			want: "Crime",
		},
		{
			name: "get primitive value from slice with filter using attribute name (error)",
			args: args{
				obj:       getTestStruct(),
				attribute: "GenreNames[Name='Crime']",
			},
			want: dipper.ErrFilterNotFound,
		},
		{
			name: "get struct field from slice element with filter",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres[Name=='Crime'].Name",
			},
			want: "Crime",
		},
		{
			name: "get map attribute value from slice element with filter",
			args: args{
				obj:       toJSONMap(getTestStruct()),
				attribute: "genres[name='Crime'].description",
			},
			want: "Narratives that centre on criminal acts and especially on the investigation of a crime, often a murder",
		},
		{
			name: "get map attribute value from slice element with filter by integer value",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres[ID==0].Name",
			},
			want: "Mystery",
		},
		{
			name: "get map attribute value from slice element with filter by float value",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres[ID=0.0].Name",
			},
			want: "Mystery",
		},
		{
			name: "get float value from slice with filter",
			args: args{
				obj:       []interface{}{0, 1.0, 1.5, 2, uint(3)},
				attribute: "[=1.5]",
			},
			want: 1.5,
		},
		{
			name: "get uint value from slice with filter",
			args: args{
				obj:       []interface{}{0, 1.0, 1.5, 2, uint(3)},
				attribute: "[=3]",
			},
			want: uint(3),
		},
		{
			name: "get boolean value from slice with filter",
			args: args{
				obj:       []interface{}{0.0, 1, true, false, nil},
				attribute: "[=true]",
			},
			want: true,
		},
		{
			name: "get null value from slice with filter",
			args: args{
				obj:       []interface{}{0.0, 1, true, false, nil},
				attribute: "[=null]",
			},
			want: nil,
		},
		{
			name: "invalid filter expression",
			args: args{
				obj:       getTestStruct(),
				attribute: "GenreNames['Mystery']",
			},
			want: dipper.ErrInvalidIndex,
		},
		{
			name: "invalid filter value",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres[ID={}]",
			},
			want: dipper.ErrInvalidFilterValue,
		},
		{
			name: "invalid filter expression",
			args: args{
				obj:       getTestStruct(),
				attribute: "GenreNames[*==]",
			},
			want: dipper.ErrInvalidFilterExpression,
		},
		{
			name: "separator before filter expression",
			args: args{
				obj:       getTestStruct(),
				attribute: "Genres.[ID=0].Name",
			},
			want: dipper.ErrInvalidIndex,
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

func TestDipper_SetWithFilter(t *testing.T) {
	type args struct {
		v         interface{}
		attribute string
		newValue  interface{}
	}
	type want struct {
		attribute string
		result    interface{}
		deleted   bool
		newValue  interface{}
	}
	tests := []struct {
		name      string
		separator string // Default is "."
		args      args
		want      want
	}{
		{
			name: "update string value in slice",
			args: args{
				v:         getTestStruct(),
				attribute: "GenreNames[='Mystery']",
				newValue:  "Romance",
			},
			want: want{
				result:    nil,
				attribute: "GenreNames.0",
				newValue:  "Romance",
			},
		},
		{
			name: "update struct field",
			args: args{
				attribute: "Genres[Name='Mystery'].Name",
				v:         getTestStruct(),
				newValue:  "Romance",
			},
			want: want{
				result:    nil,
				attribute: "Genres.0.Name",
				newValue:  "Romance",
			},
		},
		{
			name: "update struct field with filter by integer field",
			args: args{
				attribute: "Genres[ID=0.0].Name",
				v:         getTestStruct(),
				newValue:  "Romance",
			},
			want: want{
				result:    nil,
				attribute: "Genres.0.Name",
				newValue:  "Romance",
			},
		},
		{
			name: "update map attribute",
			args: args{
				attribute: "genres[id=0].name",
				v:         toJSONMap(getTestStruct()),
				newValue:  "Romance",
			},
			want: want{
				result:    nil,
				attribute: "genres.0.name",
				newValue:  "Romance",
			},
		},
		{
			name: "update map attribute with filter by float value",
			args: args{
				attribute: "genres[id=0.0].name",
				v:         toJSONMap(getTestStruct()),
				newValue:  "Romance",
			},
			want: want{
				result:    nil,
				attribute: "genres.0.name",
				newValue:  "Romance",
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
				newValue := d.Get(tt.args.v, tt.want.attribute)
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
