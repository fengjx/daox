package daox

import (
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"testing"
	"time"
)

type testInfo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func TestBatchFetch(t *testing.T) {
	redisCtl := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	cacheTool := NewCacheTool(redisCtl, time.Minute*10)
	var infos []*testInfo
	err := cacheTool.BatchFetch("test-batch", []string{"1", "2", "3"}, &infos, func(missItem []string) (map[string]interface{}, error) {
		res := make(map[string]interface{}, 0)
		for _, item := range missItem {
			i, _ := strconv.Atoi(item)
			res[item] = &testInfo{
				Id:   int64(i),
				Name: fmt.Sprintf("name-%s", item),
			}
		}
		return res, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	jsonStr, err := json.Marshal(infos)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsonStr))
}
