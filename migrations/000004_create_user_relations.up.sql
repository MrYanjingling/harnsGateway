CREATE TABLE IF NOT EXISTS `user_roles` (
  `user_id` CHAR(32) NOT NULL,
  `role_id` CHAR(32) NOT NULL,
  `created_time` DATETIME NOT NULL,
  PRIMARY KEY (`user_id`,`role_id`),
  KEY `idx_user_roles_role` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-角色关联';

CREATE TABLE IF NOT EXISTS `user_orgs` (
  `user_id` CHAR(32) NOT NULL,
  `org_id` CHAR(32) NOT NULL,
  `created_time` DATETIME NOT NULL,
  PRIMARY KEY (`user_id`,`org_id`),
  KEY `idx_user_orgs_org` (`org_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-部门关联';
