create table if not exists sys_user
(
    `id`       bigint auto_increment primary key,
    `username` varchar(32)  not null comment '用户名',
    `pwd`      varchar(64)  not null comment '密码',
    `salt`     varchar(64)  not null comment '密码盐',
    `email`    varchar(64)  not null default '' comment '邮箱',
    `nickname` varchar(32)  not null comment '昵称',
    `avatar`   varchar(256) not null default '' comment '头像',
    `phone`    varchar(32)  not null default '' comment '手机号',
    `status`   varchar(32)  not null default 'normal' comment '用户状态',
    `remark`   varchar(512) not null default '' comment '备注',
    `utime`    timestamp    not null default current_timestamp on update current_timestamp comment '更新时间',
    `ctime`    timestamp    not null default current_timestamp comment '创建时间',
    unique uk_u (`username`),
    index idx_u_s (`username`, `status`)
)
    engine = innodb
    default charset = utf8mb4 comment '用户信息表';