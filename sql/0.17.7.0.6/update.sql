insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('disaster_recovery.sh','oes灾备恢复','$1:集群号名称|字符串|必填|示例:01','oes','cmd_emgy_oes','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('disaster_recovery.sh','mds灾备恢复','$1:集群号名称|字符串|必填|示例:01','mds','cmd_emgy_mds','shell','1','1','ansible');
delete from job_script WHERE name = 'oes_colony_restart.sh' and project = 'oes' and label = 'cmd_emgy_oes';
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('stk_colony_restart.sh','现货集群拉取上场文件重启','$1:集群号名称|字符串|必填|示例:01','oes','cmd_emgy_oes','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('crd_colony_restart.sh','两融集群拉取上场文件重启','$1:集群号名称|字符串|必填|示例:01','oes','cmd_emgy_oes','shell','1','1','ansible');
insert into job_script(name,descr,param_desc,project,label,language,status,is_builtin,username) values('opt_colony_restart.sh','期权集群拉取上场文件重启','$1:集群号名称|字符串|必填|示例:01','oes','cmd_emgy_oes','shell','1','1','ansible');
