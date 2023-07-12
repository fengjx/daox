# daox

基于 sqlx + go-redis 的轻量级数据库访问辅助工具，daox 的定位是 sqlx 的功能增强，不是一个 orm。

封装了基础 crud api，同时 `sqlbuilder` 也能够帮助构造 sql。

实现了代码生成器，有内置生成模板，也可以自定义模板。

为什么有 orm 框架了还要开发 daox？

> `orm` 框架使用简单，但往往底层实现比较复杂，对大多数人来说很难二次开发扩展，并且在大多数情况下我们只使用了`orm`框架可能不到 20% 的功能。
> `daox` 只包含基础 crud api，提供`sqlbuilder`，代码简单，容易做二次扩展。

## 安装

```
go get github.com/fengjx/daox
```

## CRUD

```sql
-- 在MySQL创建测试表
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
```

查询
```go

```

修改
```go

```

删除
```go

```


缓存
```go

```

## sqlbuilder








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


