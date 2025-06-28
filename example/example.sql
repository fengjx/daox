-- Active: 1740329404403@@192.168.6.121@3306@demo
create table if not exists app_user (
    `id` bigint auto_increment primary key,
    `username` varchar(32) not null comment '用户名',
    `pwd` varchar(64) not null comment '密码',
    `salt` varchar(64) not null comment '密码盐',
    `email` varchar(64) not null default '' comment '邮箱',
    `nickname` varchar(32) not null comment '昵称',
    `avatar` varchar(256) not null default '' comment '头像',
    `phone` varchar(32) not null default '' comment '手机号',
    `status` varchar(32) not null default 'normal' comment '用户状态',
    `remark` varchar(512) not null default '' comment '备注',
    `utime` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
    `ctime` timestamp not null default current_timestamp comment '创建时间',
    unique uk_u (`username`),
    index idx_u_s (`username`, `status`)
) engine = innodb default charset = utf8mb4 comment '用户信息表';

create table if not exists app_order (
    `id` bigint auto_increment primary key,
    `user_id` bigint not null comment '用户ID',
    `amount` decimal(10, 2) not null comment '订单金额',
    `status` varchar(32) not null default 'normal' comment '订单状态',
    `remark` varchar(512) not null default '' comment '备注',
    `utime` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
    `ctime` timestamp not null default current_timestamp comment '创建时间',
    unique uk_u (`user_id`),
    index idx_u_s (`user_id`, `status`)
) engine = innodb default charset = utf8mb4 comment '订单信息表';

create table if not exists app_card (
    `id` bigint auto_increment primary key,
    `user_id` bigint not null comment '用户ID',
    `number` varchar(32) not null comment '卡号',
    `status` varchar(32) not null default 'normal' comment '卡状态',
    `utime` timestamp not null default current_timestamp on update current_timestamp comment '更新时间',
    `ctime` timestamp not null default current_timestamp comment '创建时间',
    unique uk_u (`user_id`),
    index idx_u_s (`user_id`, `status`)
) engine = innodb default charset = utf8mb4 comment '身份证信息表';

