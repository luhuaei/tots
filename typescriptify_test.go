package tots

import (
	"fmt"
	"testing"
)

func TestTots(t *testing.T) {
	type PageParameter struct {
		Page     uint        `json:"page"`
		Size     uint        `json:"size"`
		Sort     string      `json:"sort"`
		Order    string      `json:"order"`
		Fields   string      `json:"fields"`
		Filters  interface{} `json:"filters"`
		Filters2 interface{} `json:"filters2"`
		Filters3 interface{} `json:"filters3"`
		Keyword  string      `json:"keyword" ts_doc:"搜索关键词字段"`
	}

	converter := New().Debug().Add(PageParameter{})
	str, err := converter.Convert()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(str)
}

type fooType string

const (
	fooTypeA = "a"
	fooTypeB = "b"
)

func TestUnionTsType(t *testing.T) {
	r := UnionTsType("a", "b", "c", "d")
	fmt.Println(r)
	if r != `"a"|"b"|"c"|"d"` {
		panic("union incorrect")
	}

	r1 := UnionTsType(1, 2, 3, 4)
	fmt.Println(r1)
	if r1 != `1|2|3|4` {
		panic("union incorrect")
	}

	r2 := UnionTsType(1.1, 2.2, 3.3, 4.4)
	fmt.Println(r2)
	if r2 != `1.1|2.2|3.3|4.4` {
		panic("union incorrect")
	}

	r3 := UnionTsType(fooTypeA, fooTypeB)
	fmt.Println(r3)
	if r3 != `"a"|"b"` {
		panic("union incorrect")
	}
}
