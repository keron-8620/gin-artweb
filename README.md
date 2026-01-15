####################################### Web程序部署 ###########################################################
1. 安装依赖环境
yum -y install gd gd-devel gcc gcc-c++ krb5-devel libffi-devel openssl-devel zlib-devel xz-devel rsync pcre pcre-devel at python3-devel sshpass mariadb mariadb-server mariadb-devel


2. 安装python3.8的依赖环境
注意：请使用ansible-install_v5.9.tar.gz包

3. 修改nginx配置文件nginx/conf/nginx.conf配置文件中的mon的ip和端口(不用mon嵌套可忽略)

4. 修改api接口配置文件api/config.yml中的数据库相关配置

5. 生成迁移文件: python3.8 manage.py makemigrations

6. 创建数据库表: python3.8 manage.py migrate

7. 将数据导入数据库: python3.8 manage.py loaddata database.json

8. 执行start.sh 启动系统(默认用户名: mon, 密码: Quant360@mon)

9. 执行stop.sh 关闭系统

注意事项：
1. 上传程序包需要自行依据版本拷贝conf文件夹，并将对应的修改文件放入对应的文件夹内
2. 注意检查配置文件程序包内的文件权限，需要有读写权限
3. artweb需要每天重启，请自行添加计划任务
####################################### Web程序部署 ###########################################################


####################################### Web程序升级 ###########################################################
前置工作：
上传程序包到对应的服务器与原程序同级目录下，关闭源程序并重命名，然后解压程序包并重命名为原包名，并执行切换到api目录下

1. 备份数据库
python3.8 version.py -p 原程序包的api的config.yml文件路径 -o backup

2. 升级前准备，拷贝对源路径下的缓存文件到目标路径下
python3.8 version.py -p 原程序包的api的config.yml文件路径 -o before_upgrade

3. 更新数据库迁移文件(若数据库无更新，请跳过此步骤)
python3.8 manage.py makemigrations

4. 更新数据库表结构(若数据库无更新，请跳过此步骤)
python3.8 manage.py migrate

5. 升级后操作
python3.8 version.py -p 原程序包的api的config.yml文件路径 -o after_upgrade

6. 系统回退
python3.8 version.py -p 原程序包的api的config.yml文件路径 -o rollback
然后将原程序包重命名回来即可
####################################### Web程序升级 ###########################################################
