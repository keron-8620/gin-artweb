################################ Web程序部署 ##################################

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
   初始管理员凭据必须通过部署流程安全下发，禁止使用或记录默认密码。
   JWT_ACCESS_SECRET、JWT_REFRESH_SECRET 必须分别配置为不同的至少32字节随机值。
   启用 metrics 或 pprof 时，还必须设置 DIAGNOSTICS_TOKEN。
   注: 系统默认的脚本使用python3解释器，建议在虚拟环境

################################ Web程序部署 ##################################

################################ Web程序升级 ##################################

1. 依据自身的数据库类型，全量备份数据库
2. 解压程序包，将老版本下的storage文件夹拷贝到新版本下
3. 修改配置文件(配置文件可能存在变更，请参考具体的更新说明)
4. 更新数据库表结构 ./bin/artweb -migrate
5. 执行升级sql脚本
   ./bin/artweb -exec-sql sql/版本号/update.sql
6. 重启服务并登陆

本次 mon 一次性升级：先停止 art-web，直接提供本地 mon 和 JDK 程序包路径：
`sh scripts/upgrade/mon-upgrade.sh --mon-package /绝对路径/mon.tar.gz --jdk-package /绝对路径/jdk.tar.gz --yes`
脚本按手动上传程序包的规则生成随机存储文件名，创建 `mon` 和 `jdk` 程序包记录，并将所有已有 `mon_node` 关联到这两个新程序包。脚本会备份数据库、storage、旧配置，并保留原有 mon 节点 ID 及外键引用。
非 SQLite 数据库还必须设置 `ARTWEB_DB_BACKUP_COMMAND`，命令应将全量备份写入该变量指定的 `ARTWEB_DB_BACKUP_FILE`。当前非 SQLite 数据库需要另外提供 mon 节点 ID 查询能力后再执行文件目录迁移；SQLite 可直接执行完整流程。
程序包版本默认取文件名去掉最后扩展名，也可通过 `--mon-version` 和 `--jdk-version` 显式指定。升级失败时不要删除备份目录，先使用备份恢复数据库和 `storage`。

回滚：回滚数据库，关闭新版本的服务，启动老版本

################################ Web程序升级 ##################################
