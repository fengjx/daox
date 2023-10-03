package daox

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testInfo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func TestFetch(t *testing.T) {
	redisCtl := createMockRedisClient(t)
	cacheTool := NewCacheProvider(redisCtl, "fetch-test-user", "v1", time.Minute*10)
	tinfo := &testInfo{
		Id:   1,
		Name: "name-v1-1",
	}
	info := &testInfo{}
	err := cacheTool.Fetch("user-by-id", "1", info, func() (interface{}, error) {
		info = &testInfo{}
		id, _ := strconv.Atoi("1")
		info.Id = int64(id)
		info.Name = fmt.Sprintf("name-v1-%d", id)
		return info, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	testJsonStr, _ := json.Marshal(tinfo)
	jsonStr, _ := json.Marshal(info)
	t.Log(string(jsonStr))
	assert.Equal(t, string(testJsonStr), string(jsonStr))
}

func TestBatchFetch(t *testing.T) {
	redisCtl := createMockRedisClient(t)
	cache := NewCacheProvider(redisCtl, "test-cache", "v1", time.Minute*10)
	for i := 0; i < 10; i++ {
		var testInfos []*testInfo
		var items []interface{}
		for j := 0; j < 3; j++ {
			testInfos = append(testInfos, &testInfo{
				Id:   int64(j),
				Name: fmt.Sprintf("name-%d", j),
			})
			items = append(items, int64(j))
		}
		var infos []*testInfo
		err := cache.BatchFetch("test-batch", items, &infos, func(missItem []interface{}) (map[interface{}]interface{}, error) {
			res := make(map[interface{}]interface{}, 0)
			for _, item := range missItem {
				res[item] = &testInfo{
					Id:   item.(int64),
					Name: fmt.Sprintf("name-%v", item),
				}
			}
			return res, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		infoMap := make(map[int64]*testInfo)
		for _, info := range infos {
			infoMap[info.Id] = info
		}

		for _, info := range testInfos {
			testJsonStr, _ := json.Marshal(info)
			jsonStr, _ := json.Marshal(infoMap[info.Id])
			assert.Equal(t, string(testJsonStr), string(jsonStr))
		}
	}
}
