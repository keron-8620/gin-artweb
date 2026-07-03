DELETE FROM sys_role_api WHERE api_id IN ('2018', '5031', '5032', '5033', '5034', '5035');

DELETE FROM sys_menu_api WHERE api_id=2018;

DELETE FROM sys_api WHERE id IN ('2018', '5031', '5032', '5033', '5034', '5035');

DELETE FROM job_script WHERE project="oes" AND label = "agw";

update sys_api set url='/api/v1/mds/:colony_num/conf' WHERE id=4021;
update sys_api set url='/api/v1/mds/:colony_num/conf/:dir_name' WHERE id=4022;
update sys_api set url='/api/v1/mds/:colony_num/conf/:dir_name/:filename' WHERE id=4025;
update sys_api set url='/api/v1/mds/:colony_num/conf/:dir_name/:filename' WHERE id=4026;

update sys_api set url='/api/v1/oes/:colony_num/conf' WHERE id=5021;
update sys_api set url='/api/v1/oes/:colony_num/conf/:dir_name' WHERE id=5022;
update sys_api set url='/api/v1/oes/:colony_num/conf/:dir_name/:filename' WHERE id=5025;
update sys_api set url='/api/v1/oes/:colony_num/conf/:dir_name/:filename' WHERE id=5026;
