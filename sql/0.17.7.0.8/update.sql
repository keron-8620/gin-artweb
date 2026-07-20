insert into sys_api(id,url,method,label,descr) values(5041,'/api/v1/oes/agw/conf/:id','GET','oes','查询agw配置文件列表');
insert into sys_api(id,url,method,label,descr) values(5042,'/api/v1/oes/agw/conf/:id','POST','oes','上传agw配置文件');
insert into sys_api(id,url,method,label,descr) values(5045,'/api/v1/oes/agw/conf/:id','DELETE','oes','删除agw配置文件');
insert into sys_api(id,url,method,label,descr) values(5046,'/api/v1/oes/agw/conf/:id/download','GET','oes','下载agw配置文件');


insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values(8,'/','/','/','{"title":"/","icon":""}',8,1,'/',null);
insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values(9,'dashboard','dashboard','dashboard','{"title":"首页","icon":""}',9,1,'首页',8);
insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values(64,'agw','agw','Agw','{"title":"agw","icon":""}',5,1,'agw',2);
insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values(65,'agw_conf','agw_conf','agw_conf','{"title":"agw配置文件","icon":""}',6,1,'agw配置文件',2);



insert into sys_menu_api(menu_id,api_id) values(9,4006);
insert into sys_menu_api(menu_id,api_id) values(9,5006);
insert into sys_menu_api(menu_id,api_id) values(9,5007);
insert into sys_menu_api(menu_id,api_id) values(9,5008);
insert into sys_menu_api(menu_id,api_id) values(64,1001);
insert into sys_menu_api(menu_id,api_id) values(64,1011);
insert into sys_menu_api(menu_id,api_id) values(64,2001);
insert into sys_menu_api(menu_id,api_id) values(64,2012);
insert into sys_menu_api(menu_id,api_id) values(64,5031);
insert into sys_menu_api(menu_id,api_id) values(64,5032);
insert into sys_menu_api(menu_id,api_id) values(64,5033);
insert into sys_menu_api(menu_id,api_id) values(64,5034);
insert into sys_menu_api(menu_id,api_id) values(64,5035);
insert into sys_menu_api(menu_id,api_id) values(65,5031);
insert into sys_menu_api(menu_id,api_id) values(65,5041);
insert into sys_menu_api(menu_id,api_id) values(65,5042);
insert into sys_menu_api(menu_id,api_id) values(65,5045);
insert into sys_menu_api(menu_id,api_id) values(65,5046);



insert into sys_role_api(role_id,api_id) values(1,5041);
insert into sys_role_api(role_id,api_id) values(1,5042);
insert into sys_role_api(role_id,api_id) values(1,5045);
insert into sys_role_api(role_id,api_id) values(1,5046);


insert into sys_role_menu(role_id,menu_id) values(1,8);
insert into sys_role_menu(role_id,menu_id) values(1,9);
insert into sys_role_menu(role_id,menu_id) values(1,64);
insert into sys_role_menu(role_id,menu_id) values(1,65);