<h3 align="center"><b>dipper</b></h3>
<p align="center">A Go library to get and set (almost) anything using simple notation</p>

<p align="center">
  <a href="https://github.com/flusflas/dipper/actions/workflows/ci.yaml"><img src="https://github.com/flusflas/dipper/actions/workflows/ci.yaml/badge.svg" alt="CI Workflow"></a>
  <a href="https://goreportcard.com/report/github.com/flusflas/dipper"><img src="https://goreportcard.com/badge/github.com/flusflas/dipper" alt="Go Report Card"></a>
  <a href="https://codecov.io/gh/flusflas/dipper"><img src="https://codecov.io/gh/flusflas/dipper/branch/master/graph/badge.svg" alt="codecov"></a>
  <img src="https://img.shields.io/badge/go%20version-%3E=1.13-F37F40.svg" alt="Go Version">
  <a href="https://github.com/flusflas/dipper/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License"></a>
  <a href="https://pkg.go.dev/github.com/flusflas/dipper"><img src="https://pkg.go.dev/badge/github.com/flusflas/dipper" alt="Go Documentation"></a>
</p>

## Install

```shell
go get github.com/flusflas/dipper
```

## What is *dipper*?

**dipper** is a simple library that let you use dot notation (or any other
delimiter-separated attribute notation) to access values in an object, both for
getting and setting, even if they are deeply nested. You can use it with
structs, maps and slices.

## How to use it?

You can create a `Dipper` instance to customize the access options:

```go
library := Library{
    Address: "123 Fake Street",
    Books: map[string]Book{
        "Dune": {
            Title:  "Dune",
            Year:   1965,
            Genres: []string{"Novel", "Science fiction", "Adventure"},
        },
        "Il nome della rosa": {
            Title:  "Il nome della rosa",
            Year:   1980,
            Genres: []string{"Novel", "Mystery"},
        },
    },
}

d := dipper.New(dipper.Options{Separator: "->"})

book := d.Get(library, "Books->Dune")
if err := dipper.Error(book); err != nil {
    return err
}
```

Or you can use the default functions if you just need dot notation access.
This is an example of how to get a nested value from an object:

```go

field := dipper.Get(library, "Books.Dune.Genres.1")  // "Science fiction"
if err := dipper.Error(field); err != nil {
    return err
}
``` 

You can also get multiple attributes at once:

```go
fields := dipper.GetMany(library, []string{
    "Address",
    "Books.Il nome della rosa.Year",
    "Books.Dune.Author",
})

// fields => map[string]interface{}{
//   "Address":                       "123 Fake Street",
//   "Books.Il nome della rosa.Year": 1980,
//   "Books.Dune.Author":             dipper.ErrNotFound,
// }

if err := fields.FirstError(); err != nil {
    return err  // Returns "dipper: not found"
}
``` 

Finally, you can also set values in addressable objects:

```go
err := dipper.Set(&library, "Books.Dracula", Book{Year: 1897})
``` 

There are two special values that can be used in `Set()`:
- `Zero`, to set the attribute to its zero value.
- `Delete`, to delete a map key. If the attribute is not a map value, the value
  will be zeroed.


## Notes

- This library works with reflection. It has been designed to have a good
  trade-off between features and performance.
- The only supported type for map keys is `string`. `map[interface{}]` is also
  allowed if the underlying value is a `string`.
- Errors are not returned explicitly in `Get()` and `GetMany()` to support
  accessing multiple attributes at a time and getting a clear result. Instead,
  error handling functions are provided.
- Struct fields have to be exported, both for getting and setting. Trying to
  access an unexported struct field will return `ErrUnexported`.
- Using maps with keys containing your Dipper delimiter (or `.` if using the
  convenience functions) is not supported for obvious reasons. If you're trying
  to access a map with conflicting characters, use a custom `Dipper` with a
  different field separator.

### Future ideas

- Case sensitivity option.
- Tag option for struct fields.
- Attribute expansion (e.g. `Books.*.Title`).
- Custom object parser.
- Option to access unexported fields.
