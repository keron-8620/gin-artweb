delete from job_script where name = 'disaster_recovery.sh' and project = 'oes' and label = 'cmd_emgy_oes';
delete from job_script where name = 'disaster_recovery.sh' and project = 'mds' and label = 'cmd_emgy_mds';

delete from job_script where name = 'stk_colony_restart.sh' and project = 'oes' and label = 'cmd_emgy_oes';
delete from job_script where name = 'crd_colony_restart.sh' and project = 'oes' and label = 'cmd_emgy_oes';
delete from job_script where name = 'opt_colony_restart.sh' and project = 'oes' and label = 'cmd_emgy_oes';

insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('oes_colony_restart.sh','oes集群拉取上场文件重启','$1:集群号名称|字符串|必填|示例:01','oes','cmd_emgy_oes','shell','1','1','ansible');
