package main

import (
	"context"
	"fmt"

	"github.com/rickslab/ares/es"
	_ "github.com/rickslab/ares/logger"
	"github.com/rickslab/ares/util"
)

func main() {
	ctx := context.Background()
	cli := es.Client()

	/*res, err := cli.Create(ctx, "athena-task", 126, es.Object{
		"name":    "test5",
		"content": "这是一条测试任务",
	})
	util.AssertError(err)

	res, err := cli.Delete(ctx, "athena-task", 124)
	util.AssertError(err)

	res, err := cli.Update(ctx, "athena-task", 126, es.Object{
		"doc": es.Object{
			"name": "测试5",
		},
	})
	util.AssertError(err)*/

	res, err := cli.Search(ctx, "athena-task", es.Object{
		"query": es.Object{
			"match": es.Object{
				"content": "测试",
			},
		},
		"highlight": es.Object{
			"pre_tags":  []string{"<font color='red'>"},
			"post_tags": []string{"</font>"},
			"fields": es.Object{
				"content": es.Object{},
			},
		},
	}, 0, 10)
	util.AssertError(err)

	fmt.Printf("res=%+v\n", *res)
}
