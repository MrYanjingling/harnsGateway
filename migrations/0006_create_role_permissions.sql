CREATE TABLE IF NOT EXISTS `role_permissions` (
  `role_id` CHAR(32) NOT NULL,
  `permission` VARCHAR(64) NOT NULL,
  PRIMARY KEY (`role_id`,`permission`),
  KEY `idx_role_permissions_permission` (`permission`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限';
