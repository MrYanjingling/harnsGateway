CREATE TABLE IF NOT EXISTS `roles` (
  `id` CHAR(32) NOT NULL COMMENT 'ULID',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
  `name` VARCHAR(64) NOT NULL COMMENT '角色名',
  `code` VARCHAR(64) NOT NULL COMMENT '编码',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `status` VARCHAR(16) NOT NULL DEFAULT 'enabled' COMMENT '状态 enabled/disabled',
  `created_user` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_user` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
  `created_time` DATETIME NOT NULL COMMENT '创建时间',
  `updated_time` DATETIME NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_roles_name` (`name`),
  UNIQUE KEY `uk_roles_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色';
