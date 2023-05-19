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
	redisCtl := createRedisClient(t)
	//redisCtl = redis.NewClient(&redis.Options{
	//	Addr: "127.0.0.1:6379",
	//})
	cacheTool := NewCacheTool(redisCtl, time.Minute*10)
	tinfo := &testInfo{
		Id:   1,
		Name: "name-v1-1",
	}
	info := &testInfo{}
	err := cacheTool.Fetch("test-fetch-v1", "1", info, func(missItem string, dest interface{}) error {
		res := dest.(*testInfo)
		id, _ := strconv.Atoi(missItem)
		res.Id = int64(id)
		res.Name = fmt.Sprintf("name-v1-%d", id)
		return nil
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
	redisCtl := createRedisClient(t)
	cacheTool := NewCacheTool(redisCtl, time.Minute*10)
	for i := 0; i < 10; i++ {
		var testInfos []*testInfo
		var items []string
		for j := 0; j < 3; j++ {
			testInfos = append(testInfos, &testInfo{
				Id:   int64(j),
				Name: fmt.Sprintf("name-%d", j),
			})
			items = append(items, fmt.Sprintf("%d", j))
		}
		var infos []*testInfo
		err := cacheTool.BatchFetch("test-batch-v2", items, &infos, func(missItem []string) (map[string]interface{}, error) {
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
