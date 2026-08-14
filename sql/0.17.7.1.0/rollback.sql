DELETE FROM sys_role_menu WHERE menu_id IN (51, 52);

DELETE FROM sys_menu_api WHERE menu_id IN (51, 52);

DELETE FROM sys_menu WHERE id IN (51, 52);


DELETE FROM sys_role_api WHERE api_id IN (3011, 3012, 3015, 3016);

DELETE FROM sys_api WHERE id IN (3011, 3012, 3015, 3016);


DELETE FROM job_script WHERE project = 'mon' and label = 'dep' AND name = 'deploy.sh';

UPDATE sys_menu SET meta = '{"title":"AGW","icon":""}', descr='agw' WHERE id = 51;
UPDATE sys_menu SET meta = '{"title":"AGW配置文件","icon":""}', descr='agw配置文件' WHERE id = 65;

