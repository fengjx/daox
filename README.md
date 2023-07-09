# daox

基于 sqlx + go-redis 的轻量级数据库访问辅助工具，daox 的定位是 sqlx 的功能增强，不是一个 orm。

封装了基础 crud api，同时 `sqlbuilder` 也能够帮助构造 sql。

实现了代码生成器，有内置生成模板，也可以自定义模板。

## 安装

```
go get github.com/fengjx/daox
```

## CRUD

创建 dao 对象
```go
db, err := mysqlDB()
if err != nil {
    return nil, err
}
// 初始化 sqlx
dbx := sqlx.NewDb(db, "mysql")
dbx.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
// 初始化 daox，绑定表`user`，主键字段名`id`，主键为自增 id
dao := daox.NewDAO(db, "user", "id", reflect.TypeOf(&User{}), daox.IsAutoIncrement())
```



### 新增


### 修改


### 删除


### 查询




## 缓存






## 代码生成

### 安装代码生成工具

```bash
go install github.com/fengjx/daox/cmd/gen@latest
gen -h

GLOBAL OPTIONS:
   -f value    config file path
   --help, -h  show help
```

生成代码

```
gen -f gen.yml
```

配置示例说明

```yaml
ds:
  type: mysql
  dsn: root:1234@tcp(192.168.1.200:3306)/gogo
target:
  custom:
    tag-name: db
    out-dir: ./out
    template-dir:
    var:
      a: aa
      b: bb
  tables:
    - user
    - blog
```

自定义模板说明


