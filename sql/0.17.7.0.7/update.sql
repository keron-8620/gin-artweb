insert into sys_api(id,url,method,label,descr) values('5031','/api/v1/oes/agw','GET','oes','查询agw列表');
insert into sys_api(id,url,method,label,descr) values('5032','/api/v1/oes/agw','POST','oes','新增agw');
insert into sys_api(id,url,method,label,descr) values('5033','/api/v1/oes/agw/:id','GET','oes','查询单个agw');
insert into sys_api(id,url,method,label,descr) values('5034','/api/v1/oes/agw/:id','PUT','oes','修改单个agw');
insert into sys_api(id,url,method,label,descr) values('5035','/api/v1/oes/agw/:id','DELETE','oes','删除单个agw');

insert into sys_api(id,url,method,label,descr) values('2018','/api/v1/jobs/record/:id/log/stream','GET','job','获取脚本执行日志流');

update sys_api set url='/api/v1/mds/conf/:colony_num' WHERE id=4021;
update sys_api set url='/api/v1/mds/conf/:colony_num' WHERE id=4022;
update sys_api set url='/api/v1/mds/conf/:colony_num' WHERE id=4025;
update sys_api set url='/api/v1/mds/conf/:colony_num/download' WHERE id=4026;

update sys_api set url='/api/v1/oes/conf/:colony_num' WHERE id=5021;
update sys_api set url='/api/v1/oes/conf/:colony_num' WHERE id=5022;
update sys_api set url='/api/v1/oes/conf/:colony_num' WHERE id=5025;
update sys_api set url='/api/v1/oes/conf/:colony_num/download' WHERE id=5026;

insert into sys_menu_api(menu_id,api_id) values(84,2018);
insert into sys_menu_api(menu_id,api_id) values(80,2018);
insert into sys_menu_api(menu_id,api_id) values(83,2018);
insert into sys_menu_api(menu_id,api_id) values(81,2018);
insert into sys_menu_api(menu_id,api_id) values(85,2018);
insert into sys_menu_api(menu_id,api_id) values(86,2018);



insert into sys_role_api(role_id,api_id) values('1','5031');
insert into sys_role_api(role_id,api_id) values('1','5032');
insert into sys_role_api(role_id,api_id) values('1','5033');
insert into sys_role_api(role_id,api_id) values('1','5034');
insert into sys_role_api(role_id,api_id) values('1','5035');

insert into sys_role_api(role_id,api_id) values('1','2018');

insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('deploy.sh','部署agw','$1:编号|int|必填|示例:1','oes','agw','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('backup.sh','备份agw','$1:编号|int|必填|示例:1','oes','agw','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('rollback.sh','回滚agw','$1:编号|int|必填|示例:1','oes','agw','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('start.sh','启动agw','$1:编号|int|必填|示例:1','oes','agw','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('stop.sh','停止agw','$1:编号|int|必填|示例:1','oes','agw','shell','1','1','ansible');
