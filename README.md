# daox

基于 sqlx 的轻量级数据库访问辅助工具，daox 的定位是 sqlx 的功能增强，不是一个 orm。

封装了基础 crud api，实现了`sqlbuilder`，通过 api 生成 sql，无需手动拼接。

实现了代码生成器，有内置生成文件模板，也可以自定义模板。

## 特性

- 轻量级设计,专注于 SQL 操作的简化
- 支持读写分离
- 支持事务操作
- 支持自动生成代码
- 支持自定义字段映射
- 内置 SQL 构建器
- 支持 Context 传递
- 支持自定义 Hook

## 安装

```bash
go get github.com/fengjx/daox
```

## 文档&示例

[GoDoc](https://pkg.go.dev/github.com/fengjx/daox)

### CRUD

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
// 1. 使用全局默认DB
db := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3306)/demo")
db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
// 注册全局DB
daox.UseDefaultMasterDB(db)
dao := daox.NewDao[*User](tableName, "id", daox.IsAutoIncrement())

// 2. 使用 NewDao 创建,指定主从库
masterDB := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3306)/demo")
readDB := sqlx.MustOpen("mysql", "root:1234@tcp(localhost:3307)/demo")
masterDB.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
readDB.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
dao := daox.NewDao[*User](tableName, "id", 
    daox.IsAutoIncrement(),
    daox.WithDBMaster(masterDB),
    daox.WithDBRead(readDB),
)

// 3. 使用 meta 接口创建
dao := daox.NewDaoByMeta(UserMeta)
```

新增
```go
ctx := context.Background()

// 单条插入
user := &User{
    Uid:      1001,
    Nickname: "test",
    Sex:      1,
    Utime:    time.Now().Unix(),
    Ctime:    time.Now().Unix(),
}
id, err := dao.SaveContext(ctx, user)
if err != nil {
    log.Panic(err)
}

// 批量插入
users := make([]*User, 0)
for i := 0; i < 20; i++ {
    sec := time.Now().Unix()
    user := &User{
        Uid:      100 + int64(i),
        Nickname: randString(6),
        Sex:      int32(i) % 2,
        Utime:    sec,
        Ctime:    sec,
    }
    users = append(users, user)
}
affected, err := dao.BatchSaveContext(ctx, users)
if err != nil {
    log.Panic(err)
}

// Replace Into 插入
id, err = dao.ReplaceIntoContext(ctx, user)

// Insert Ignore 插入
id, err = dao.IgnoreIntoContext(ctx, user)
```

查询
```go
ctx := context.Background()

// id 查询
user := new(User)
exists, err := dao.GetByIDContext(ctx, 1, user)

// 批量id查询
var list []*User
err = dao.ListByIDsContext(ctx, &list, 10, 11)

// 指定字段查询单条记录
user := new(User)
exists, err := dao.GetByColumnContext(ctx, daox.OfKv("uid", 10000), user)

// 指定字段查询多条记录
var list []*User
err := dao.ListContext(ctx, daox.OfKv("sex", 0), &list)

// 指定字段查询多个值
var list []*User
err = dao.ListByColumnsContext(ctx, daox.OfMultiKv("uid", 10000, 10001), &list)
```

修改
```go
ctx := context.Background()

// 全字段更新
user := new(User)
exists, err := dao.GetByIDContext(ctx, 10, user)
user.Nickname = "update-name-10"
ok, err := dao.UpdateContext(ctx, user)

// 部分字段更新
ok, err = dao.UpdateFieldContext(ctx, 11, map[string]any{
    "nickname": "update-name-11",
})

// 条件更新
ok, err = dao.UpdateByCondContext(ctx, user, 
    sqlbuilder.C().Where(true, "age > ?").And(true, "sex = ?"),
    "create_time", // 忽略更新的字段
)
```

删除
```go
ctx := context.Background()

// 按 id 删除
ok, err := dao.DeleteByIDContext(ctx, 21)

// 按指定字段删除
affected, err := dao.DeleteByColumnContext(ctx, daox.OfKv("uid", 101))

// 按字段删除多条记录
affected, err = dao.DeleteByColumnsContext(ctx, daox.OfMultiKv("uid", 102, 103))
```

### sqlbuilder

创建Builder对象
```go
// 通过dao对象实例方法创建
dao.SQLBuilder()

// 独立使用
sqlbuilder.New("user_info")
```

构造sql
```go
// 查询
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

// 插入
inserter := sqlbuilder.New("user_info").Insert().
    Columns("username", "age", "sex")

sql, err := inserter.Sql()
//  INSERT INTO `user_info`(`username`, `age`, `sex`) VALUES (?, ?, ?);

nameSql, err := inserter.NameSql()
// INSERT INTO `user_info`(`username`, `age`, `sex`) VALUES (:username, :age, :sex);

// 更新
updateSQL, err := sqlbuilder.New("user_info").
    Update().
    Columns("username", "age").
    Where(sqlbuilder.C().Where(true, "id = ?")).
    Sql()
// UPDATE `user_info` SET `username` = ?, `age` = ? WHERE id = ?;

// 删除
deleteSQL, err := sqlbuilder.New("user_info").Delete().
    Where(sqlbuilder.C().Where(true, "id = ?")).
    Sql()
// DELETE FROM `user_info` WHERE id = ?;
```

更多示例请查看[sqlbuilder/sql_test.go](https://github.com/fengjx/daox/blob/master/sqlbuilder/sql_test.go)

### Hook

daox 支持注册 Hook 来实现 SQL 执行前后的自定义处理:

```go
// 全局 Hook
daox.RegisterHook(func(ctx context.Context, stmt string, args []any) {
    log.Printf("sql: %s, args: %v", stmt, args)
})

// 单个 dao 实例的 Hook
dao := daox.NewDao[*User](tableName, "id", 
    daox.WithHook(func(ctx context.Context, stmt string, args []any) {
        log.Printf("sql: %s, args: %v", stmt, args)
    }),
)
```

### 事务

daox 支持事务操作，示例如下:

```go
ctx := context.Background()
tx, err := dao.GetMasterDB().Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// 创建事务 dao
txDao := dao.WithExecutor(tx)

// 执行事务操作
id, err := txDao.SaveContext(ctx, user)
if err != nil {
    return err
}

affected, err := txDao.UpdateFieldContext(ctx, id, map[string]any{
    "nickname": "tx-update",
})
if err != nil {
    return err
}

return tx.Commit()
```

### 代码生成

#### 安装代码生成工具

```bash
$ go install github.com/fengjx/daox/cmd/gen@latest
$ gen -h

GLOBAL OPTIONS:
   -f value    config file path
   --help, -h  show help
```

生成代码

```bash
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
      user:
        module: sys
      blog:
        module: core
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
attr := map[string]any{
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

## License

MIT License