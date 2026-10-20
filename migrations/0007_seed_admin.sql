-- Bootstrap super administrator.
-- Default account: admin / admin@123 (please change it after first login).
INSERT IGNORE INTO `roles`
  (`id`, `version`, `name`, `code`, `description`, `status`, `created_user`, `updated_user`, `created_time`, `updated_time`)
VALUES
  ('00000000000000000000000000000001', 0, '管理员', 'admin', '内置超级管理员', 'enabled', 'system', 'system', NOW(), NOW());

INSERT IGNORE INTO `role_permissions` (`role_id`, `permission`)
VALUES ('00000000000000000000000000000001', '*');

INSERT IGNORE INTO `users`
  (`id`, `version`, `user_name`, `password`, `user_no`, `nick_name`, `birth_date`, `gender`, `mail`, `phone`, `note`,
   `auth_key`, `auth_mode`, `last_login_time`, `login_failure_times`, `locked_until`, `expire_times`, `status`,
   `created_user`, `updated_user`, `created_time`, `updated_time`)
VALUES
  ('00000000000000000000000000000002', 0, 'admin',
   '$2a$10$HD89t3VpmaKpkME25v5e/.aCwFycTmNPY6tz8WxNJZXQYqMVvNIzW',
   'admin', '管理员', NULL, '', '', '', '', '', 'local', NULL, 0, NULL, 0, 'enabled',
   'system', 'system', NOW(), NOW());

INSERT IGNORE INTO `user_roles` (`user_id`, `role_id`, `created_time`)
VALUES ('00000000000000000000000000000002', '00000000000000000000000000000001', NOW());
