CREATE TABLE IF NOT EXISTS `orgs` (
  `id` CHAR(32) NOT NULL COMMENT 'ULID',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
  `code` VARCHAR(64) NOT NULL COMMENT '组织编码',
  `name` VARCHAR(64) NOT NULL COMMENT '组织名称',
  `parent_id` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '父ID，空/0 为根',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '组织排序',
  `description` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '组织描述',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '组织备注',
  `status` VARCHAR(16) NOT NULL DEFAULT 'enabled' COMMENT '状态 enabled/disabled',
  `created_user` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_user` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
  `created_time` DATETIME NOT NULL COMMENT '创建时间',
  `updated_time` DATETIME NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orgs_code` (`code`),
  KEY `idx_orgs_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组织(部门)';
