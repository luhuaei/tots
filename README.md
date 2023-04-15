# fork `github.com/tkrajina/typescriptify-golang-structs/tscriptify`

# generate typescript interface for golang type

- adjust comment position
- remove `ts_transform` logic
- remove convert to file
- remove convert to `class`, only `interface`

# example

```go
type UserUid struct {
	Uid uint `gorm:"not null" json:"uid" ts_doc:"用户的id"`
}
type SocialAttribute struct {
	GORMModel
	UserUid

	Follows []uint `gorm:"type:integer[]" json:"follows"`
	Fans    []uint `gorm:"type:integer[]" json:"fans"`
}
```

```go
package main
import ""
func exportTypescript() {
	converter := tots.
		New().
		ManageType(time.Time{}, tots.TypeOptions{TSType: "string"}).
		Add(model.SocialAttribute{})
	str, err := converter.Convert()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(str)
}
func main() {
    exportTypescript()
}
```

will export
```
/* Do not change, this code is generated from Golang structs */


export interface SocialAttribute {
    id: number;
    createdAt: string;
    updatedAt: string;
    /**
     *
     * 用户的id
     */
    uid: number;
    follows: number[];
    fans: number[];
}
```

# TODO
- []use `validate` judge generate field type
