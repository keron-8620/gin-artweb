################################ Web程序部署 ###################################
1. 安装依赖环境
yum -y install rsync python3
pip3 install ansible ansible_runner
注: python版本任意，但需要ansible支持被控机的python版本
若当前环境存在多个python的版本，请在本地创建虚拟环境后再执行pip安装依赖
创建虚拟环境: python3 -m venv .venv && source .venv/bin/activate

2. 编辑config下的主配置文件system.yaml
数据库部分请依据需求自行安装，并创建指定的库

3. 创建数据库的表结构
./bin/artweb -init-database

4. 导入sql脚本
./bin/artweb -exec-sql sql/database.sql

5. 启动bin目录下的可执行程序，通过浏览器访问页面
默认的用户名为mon, 密码为Quant360@mon
注: 系统默认的脚本使用python3解释器，建议在虚拟环境中启动
################################ Web程序部署 #####################################
