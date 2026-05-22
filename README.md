\################################ Web程序部署 ##################################

1. 安装依赖环境
   yum -y install rsync python3
2. pip3 install ansible ansible\_runner
3. python版本任意，但需要ansible支持被控机的python版本
   若当前环境存在多个python的版本，请在本地创建虚拟环境后再执行pip安装依赖
   创建虚拟环境: python3 -m venv .venv && source .venv/bin/activate
4. 编辑config下的主配置文件system.yaml
   数据库部分请依据需求自行安装，并创建指定的库
5. 创建数据库的表结构
   ./bin/artweb -migrate
6. 导入sql脚本
   ./bin/artweb -exec-sql sql/database.sql
7. 启动bin目录下的可执行程序，通过浏览器访问页面
   启动命令: sh cmd/start.sh
   停止命令: sh cmd/stop.sh
   默认的用户名为mon, 密码为Quant360\@mon
   注: 系统默认的脚本使用python3解释器，建议在虚拟环境

\################################ Web程序部署 ##################################

<br />

\################################ Web程序升级 ##################################

1. 依据自身的数据库类型，全量备份数据库
2. 解压程序包，将老版本下的storage文件夹拷贝到新版本下
3. 修改配置文件(配置文件可能存在变更，请参考具体的更新说明)
4. 更新数据库表结构 ./bin/artweb -migrate
5. 执行升级sql脚本
   ./bin/artweb -exec-sql sql/版本号/update.sql
6. 重启服务并登陆

回滚：回滚数据库，关闭新版本的服务，启动老版本

\################################ Web程序升级 ##################################

