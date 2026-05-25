delete from sys_menu_api where menu_id = 83;
delete from sys_menu_api where menu_id = 62 and api_id = 2001;
delete from sys_menu_api where menu_id = 62 and api_id = 2012;
delete from sys_menu_api where menu_id = 72 and api_id = 2001;
delete from sys_menu_api where menu_id = 72 and api_id = 2012;

update sys_user set username = 'mon', password = '$2a$12$vmjs0S6AShmCBJSsXpJ2d.as4F2w0ywm5yzQmn8JLU9UTyM5qwf1i' where username = 'ansible';
