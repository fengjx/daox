package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fengjx/daox/v2/example/dao"
)

func main() {
	user, err := dao.UserDao.Preload("Card", "Orders").GetByIDContext(context.Background(), 1)
	if err != nil {
		panic(err)
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(userJson))
}
