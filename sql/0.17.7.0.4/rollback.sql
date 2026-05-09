insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values('7','/monitor','/monitor','Monitor','{"title":"监控管理","icon":""}','4','1','监控管理',null);



update sys_menu set parent_id = NULL, path='/mon', component='/mon' where id = '1';
update sys_menu set parent_id = NULL, path='/oes', component='/oes' where id = '2';
update sys_menu set parent_id = NULL, path='/mds', component='/mds' where id = '3';
update sys_menu set sort='4' where id = '4';
update sys_menu set sort='5' where id = '5';
update sys_menu set sort='6' where id = '6';

