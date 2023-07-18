# daox

基于 sqlx + go-redis 的轻量级数据库访问辅助工具，daox 的定位是 sqlx 的功能增强，不是一个 orm。

封装了基础 crud api，实现了`sqlbuilder`，能够帮助构造 sql，无需手动拼接。

实现了代码生成器，有内置生成文件模板，也可以自定义模板。


## 安装

```
go get github.com/fengjx/daox
```

## CRUD

在MySQL创建测试表
```sql
create table user_info
(
    id       bigint comment '主键',
    uid      bigint                not null,
    nickname varchar(32) default '' null comment '昵称',
    sex      tinyint     default 0 not null comment '性别',
    utime    bigint      default 0 not null comment '更新时间',
    ctime    bigint      default 0 not null comment '创建时间',
    primary key pk(id),
    unique uni_uid (uid)
) comment '用户信息表';
```

创建 dao 对象
```go
db := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3306)/demo")
db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
dao := daox.NewDAO(db, "user_info", "id", reflect.TypeOf(&User{}), daox.IsAutoIncrement())
```

新增
```go
for i := 0; i < 20; i++ {
    sec := time.Now().Unix()
    user := &User{
        Uid:      100 + int64(i),
        Nickname: randString(6),
        Sex:      int32(i) % 2,
        Utime:    sec,
        Ctime:    sec,
    }
    id, err := dao.Save(user)
    if err != nil {
        log.Panic(err)
    }
    log.Println(id)
}


```

查询
```go
// id 查询
user := new(User)
err := dao.GetByID(1, user)

// 批量id查询
var list []User
err = dao.ListByIds(&list3, 10, 11)

// 指定字段查询单条记录
user := new(User)
err := dao.GetByColumn(daox.OfKv("uid", 10000), user)

// 指定字段查询多条记录
var list []*User
err := dao.List(daox.OfKv("sex", 0), &list)

// 指定字段查询多个值
var list []*User
err = dao.ListByColumns(daox.OfMultiKv("uid", 10000, 10001), &list)
```

修改
```go
user := new(User)
err := dao.GetByID(10, user)
if err != nil {
    log.Fatal(err)
}
user.Nickname = "update-name-10"
// 全字段更新
ok, err := dao.Update(user)
if err != nil {
    log.Fatal(err)
}
log.Printf("update res - %v", ok)

// 部分字段更新
ok, err = dao.UpdateField(11, map[string]interface{}{
    "nickname": "update-name-11",
})
if err != nil {
    log.Fatal(err)
}
log.Printf("update res - %v", ok)
// 查询更新后的数据
var list []*User
err = dao.ListByIds(&list, 10, 11)
if err != nil {
    log.Fatal(err)
}
for _, u := range list {
    log.Println(u)
}
```

删除
```go
// 按 id 删除
ok, err := dao.DeleteById(21)
if err != nil {
    log.Fatal(err)
}
log.Printf("delete res - %v", ok)
user := new(User)
err = dao.GetByID(21, user)
if err != nil {
    log.Fatal(err)
}
log.Printf("delete by id res - %v", user.Id)

// 按指定字段删除
affected, err := dao.DeleteByColumn(daox.OfKv("uid", 101))
if err != nil {
    log.Fatal(err)
}
log.Printf("delete by column res - %v", affected)

// 按字段删除多条记录
affected, err = dao.DeleteByColumns(daox.OfMultiKv("uid", 102, 103))
if err != nil {
    log.Fatal(err)
}
log.Printf("multiple delete by column res - %v", affected)
```


缓存
```go
user := new(User)
// 按id查询并缓存
err := dao.GetByIDCache(10, user)
if err != nil {
    log.Fatal(err)
}
log.Printf("get by id with cache - %v", user)

// 删除缓存
err = dao.DeleteCache("id", 10)
if err != nil {
    log.Fatal(err)
}

// 按指定字段查询并缓存
cacheUser := new(User)
err = dao.GetByColumnCache(daox.OfKv("uid", 10001), cacheUser)
if err != nil {
    log.Fatal(err)
}
log.Printf("get by uid with cache - %v", cacheUser)
```

## sqlbuilder

创建Builder对象
```go
// 通过dao对象实例方法创建
dao.SQLBuilder()

// 独立使用
sqlbuilder.New("user_info")
```

构造sql
```go
querySQL, err := sqlbuilder.New("user_info").Select().
    Columns("id", "username", "age", "sex", "ctime").
    Where(
        sqlbuilder.C().
            Where(true, "age > ?").
            And(true, "sex = ?"),
    ).
    OrderBy(sqlbuilder.Desc("ctime")).
    Offset(10).
    Limit(10).Sql()
// SELECT `id`, `username`, `age`, `sex`, `ctime` FROM `user_info` WHERE age > ? AND sex = ? ORDER BY `ctime` DESC LIMIT 10 OFFSET 10;
log.Println(querySQL)


inserter := sqlbuilder.New("user_info").Insert().
    Columns("username", "age", "sex")

sql, err := inserter.Sql()
//  INSERT INTO `user_info`(`username`, `age`, `sex`) VALUES (?, ?, ?);
log.Println(sql)

nameSql, err := inserter.NameSql()
// INSERT INTO `user_info`(`username`, `age`, `sex`) VALUES (:username, :age, :sex);
log.Println(nameSql)


updateSQL, err := sqlbuilder.New("user_info").
    Update().
    Columns("username", "age").
        Where(sqlbuilder.C().Where(true, "id = ?")).
    Sql()
// UPDATE `user_info` SET `username` = ?, `age` = ? WHERE id = ?;
log.Println(updateSQL)


deleteSQL, err := sqlbuilder.New("user_info").Delete().
    Where(sqlbuilder.C().Where(true, "id = ?")).
    Sql()
// DELETE FROM `user_info` WHERE id = ?;
log.Println(deleteSQL)
```

更多示例请查看[sqlbuilder/sql_test.go](https://github.com/fengjx/daox/blob/master/sqlbuilder/sql_test.go)

## 代码生成

### 安装代码生成工具

```bash
$ go install github.com/fengjx/daox/cmd/gen@latest
$ gen -h

GLOBAL OPTIONS:
   -f value    config file path
   --help, -h  show help
```

生成代码

```
$ gen -f gen.yml
```

配置示例说明

```yaml
ds:
  type: mysql
  dsn: root:1234@tcp(localhost:3306)/demo
target:
  custom:
    tag-name: json
    out-dir: ./out
    template-dir:
    var:
      a: aa
      b: bb
  tables:
    - user_info
```

| 参数                         | 必须 | 说明                     |
|----------------------------|----|------------------------|
| ds.type                    | 是  | 数据库类型，暂时值支持 mysql      |
| ds.dsn                     | 是  | 数据库连接                  |
| target.custom.tag-name     | 是  | model 字段的 tagName      | 
| target.custom.out-dir      | 是  | 文件生成路径                 | 
| target.custom.template-dir | 否  | 自定义模板文件路径              | 
| target.custom.var          | 否  | 自定义参数，map结构，可以在模板文件中使用 | 
| target.custom.tables       | 是  | 需要生成文件的表名，list 结构      | 


自定义模板说明

通过`text/template`来渲染文件内容，模板语法不在此赘述，可自行查看参考文档。

模板中可以使用的变量，详细可以查看源码[cmd/gen/gen.go](/cmd/gen/gen.go#L172)
```go
attr := map[string]interface{}{
    "Var":     config.Target.Custom.Var,
    "TagName": config.Target.Custom.TagName,
    "Table":   table,
}
```

模板中可以使用的函数

- utils.FirstUpper: 首字母大写
- utils.FirstLower: 首字母小写
- utils.SnakeCase:  转下划线风格字符串
- utils.TitleCase:  转驼峰风格字符串
- utils.GonicCase:  转go风格驼峰字符串，user_id -> userID
- utils.LineString: 空字符串使用横线"-"代替
- SQLType2GoTypeString: sql类型转go类型字符串

```go
funcMap := template.FuncMap{
    "FirstUpper":           utils.FirstUpper,
    "FirstLower":           utils.FirstLower,
    "SnakeCase":            utils.SnakeCase,
    "TitleCase":            utils.TitleCase,
    "GonicCase":            utils.GonicCase,
    "LineString":           utils.LineString,
    "SQLType2GoTypeString": SQLType2GoTypeString,
}
```

参考`_example/gen`
