---
Title: Tags
Id: 86
SOId: 3531
---
Struct fields can have tags associated with them. These tags can be read by the `reflect` package to get custom information specified about a field by the developer.

```go
struct Account {
    Username      string `json:"username"`
    DisplayName   string `json:"display_name"`
    FavoriteColor string `json:"favorite_color,omitempty"`
}
```

In the above example, the tags are used to change the key names used by the `encoding/json` package when marshaling or unmarshaling JSON.

While the tag can be any string value, it's considered best practice to use space separated `key:"value"` pairs:

```go
struct StructName {
    FieldName int `package1:"customdata,moredata" package2:"info"`
}
```

The struct tags used with the `encoding/xml` and `encoding/json` package are used throughout the standard libarary.

<!-- TODO: example with multiple tags -->
<!-- TODO: link to json, xml chapters -->
