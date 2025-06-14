## 版本记录

### v1.0.0

- [完成基础crud和缓存实现](https://github.com/fengjx/daox/pull/1)
- [dao - 支持根据Model更新](https://github.com/fengjx/daox/pull/4)
- [fix: Model 中包含忽略字段时创建 Dao 失败](https://github.com/fengjx/daox/pull/6)
- [fix: [sqlbuilder] 分页sql offset 错误](https://github.com/fengjx/daox/pull/9)
- [dao 采用组合模式引入 sqlx](https://github.com/fengjx/daox/pull/11)
- [sqlbuilder 支持从 struct 解析字段](https://github.com/fengjx/daox/pull/13) 
- [代码生成工具](https://github.com/fengjx/daox/pull/14)
- [文档完善](https://github.com/fengjx/daox/pull/15)
- [单条数据查询返回是否存在数据](https://github.com/fengjx/daox/pull/20)
- [调整方法命名](https://github.com/fengjx/daox/pull/21)
- [支持 replace into](https://github.com/fengjx/daox/pull/26)
- [代码生成路径支持变量替换](https://github.com/fengjx/daox/pull/22)
- [where 条件支持参数拼接](https://github.com/fengjx/daox/pull/29)
- [sqlbuilder 条件构造器支持表达式方法](https://github.com/fengjx/daox/pull/32)
- [修复 not in 语句错误](https://github.com/fengjx/daox/pull/34)


### 1.0.3

- [通过meta接口创建dao](https://github.com/fengjx/daox/pull/37)
- [支持IFNULL查询](https://github.com/fengjx/daox/pull/42)

### 1.1.0

- [支持全局配置](https://github.com/fengjx/daox/pull/44)
- [ifnull 优化](https://github.com/fengjx/daox/pull/45)
- [with 方法支持表路由](https://github.com/fengjx/daox/pull/48)
- [update 支持字段 incr](https://github.com/fengjx/daox/pull/50)
- [sqlbuilder 支持db操作](https://github.com/fengjx/daox/pull/52)
- [insert 增加影响行数返回](https://github.com/fengjx/daox/pull/54)
- [中间件支持](https://github.com/fengjx/daox/pull/56)
- [middleware 重命名 hook](https://github.com/fengjx/daox/pull/58)
- [增加获取db方法](https://github.com/fengjx/daox/pull/59)
- [适配 engine 接口](https://github.com/fengjx/daox/pull/60)
- [dao 移除直接依赖 sqlx.Tx](https://github.com/fengjx/daox/pull/62)
- [获取db方法优化](https://github.com/fengjx/daox/pull/64)
- [fix 查询行数统计错误](https://github.com/fengjx/daox/pull/66)
- [全字段update支持指定where条件](https://github.com/fengjx/daox/pull/68)
- [where 支持 is null](https://github.com/fengjx/daox/pull/70)
- [支持join查询](https://github.com/fengjx/daox/pull/72)
- [支持count查询](https://github.com/fengjx/daox/pull/74)
- [alias优化](https://github.com/fengjx/daox/pull/76)
- [order by 支持 alias](https://github.com/fengjx/daox/pull/78)
- [dao.Inserter 方法使用默认insert字段](https://github.com/fengjx/daox/pull/80)
- [完善文档注释](https://github.com/fengjx/daox/pull/83)

### 1.1.1

- [fix order by 多个字段时sql拼接错误](https://github.com/fengjx/daox/pull/86)
- [fix ignore into sql 错误](https://github.com/fengjx/daox/pull/88)

