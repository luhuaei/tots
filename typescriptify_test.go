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
