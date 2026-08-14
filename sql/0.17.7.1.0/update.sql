insert into sys_api(id,url,method,label,descr) values(3011,'/api/v1/mon/conf/:id','GET','mon','查询mon配置文件列表');
insert into sys_api(id,url,method,label,descr) values(3012,'/api/v1/mon/conf/:id','POST','mon','上传mon配置文件');
insert into sys_api(id,url,method,label,descr) values(3015,'/api/v1/mon/conf/:id','DELETE','mon','删除mon配置文件');
insert into sys_api(id,url,method,label,descr) values(3016,'/api/v1/mon/conf/:id/download','GET','mon','下载mon配置文件');


insert into sys_role_api(role_id,api_id) values(1,3011);
insert into sys_role_api(role_id,api_id) values(1,3012);
insert into sys_role_api(role_id,api_id) values(1,3015);
insert into sys_role_api(role_id,api_id) values(1,3016);


insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values(51,'mon_conf','mon_conf','mon_conf','{"title":"MON配置文件","icon":""}',2,1,'MON配置文件',1);
insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values(52,'mon_pkg','mon_pkg','mon_pkg','{"title":"MON程序包","icon":""}',3,1,'MON程序包',1);

UPDATE sys_menu SET meta = '{"title":"AGW","icon":""}', descr='AGW' WHERE id = 51;
UPDATE sys_menu SET meta = '{"title":"AGW配置文件","icon":""}', descr='AGW配置文件' WHERE id = 65;

insert into sys_menu_api(menu_id,api_id) values(51,3001);
insert into sys_menu_api(menu_id,api_id) values(51,3011);
insert into sys_menu_api(menu_id,api_id) values(51,3012);
insert into sys_menu_api(menu_id,api_id) values(51,3015);
insert into sys_menu_api(menu_id,api_id) values(51,3016);
insert into sys_menu_api(menu_id,api_id) values(52,1011);
insert into sys_menu_api(menu_id,api_id) values(52,1012);
insert into sys_menu_api(menu_id,api_id) values(52,1013);
insert into sys_menu_api(menu_id,api_id) values(52,1015);
insert into sys_menu_api(menu_id,api_id) values(52,1016);


insert into sys_role_menu(role_id,menu_id) values(1,51);
insert into sys_role_menu(role_id,menu_id) values(1,52);


insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('deploy.sh','部署mon','$1:节点号|字符串|必填|示例:01','mon','dep','shell',1,1,'ansible');
