DELETE FROM sys_role_api WHERE role_id = 1 AND api_id IN (5041, 5042, 5045, 5046);

DELETE FROM sys_role_menu WHERE role_id = 1 AND menu_id IN (8, 9, 64, 65);

DELETE FROM sys_menu_api WHERE menu_id IN (8, 9, 64, 65);

DELETE FROM sys_menu WHERE id IN (8, 9, 64, 65);

DELETE FROM sys_api WHERE id IN (5041, 5042, 5045, 5046)
