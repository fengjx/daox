create table user_info
(
    id       bigint auto_increment comment '主键',
    uid      bigint                 not null comment '用户ID',
    nickname varchar(32) default '' null comment '昵称',
    sex      tinyint     default 0  not null comment '性别',
    utime    bigint      default 0  not null comment '更新时间',
    ctime    bigint      default 0  not null comment '创建时间',
    primary key pk (id),
    unique uni_uid (uid)
) comment '用户信息表';

