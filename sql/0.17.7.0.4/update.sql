insert into sys_menu(id,path,component,name,meta,sort,is_active,descr,parent_id) values('7','/monitor','/monitor','Monitor','{"title":"监控管理","icon":""}','4','1','监控管理',null);

update sys_menu set parent_id = 7, path='mon', component='mon' where id = '1';
update sys_menu set parent_id = 7, path='oes', component='oes' where id = '2';
update sys_menu set parent_id = 7, path='mds', component='mds' where id = '3';
update sys_menu set sort='5' where id = '4';
update sys_menu set sort='6' where id = '5';
update sys_menu set sort='7' where id = '6';

insert into sys_role_menu(role_id,menu_id) values('1','7');