pipeline {
    agent {
        node {
            label 'kylin10-sp3'
        }
    }

    parameters {
        string(name: 'version', description: '版本号')
    }

    options {
        timestamps() // 日志显示时间戳
        timeout(time: 30, unit: 'MINUTES') // 设置超时时长
        disableConcurrentBuilds() // 禁用并发构建
    }

    stages {
        stage('初始化环境变量') {
            steps {
                script {
                    env.version = "${params.version}"
                    env.jenkinTMP = "/tmp/jenkins-${env.BUILD_NUMBER}"
                    env.tarFileName = "artweb-${params.version}-kylin.tar.gz"
                }
            }
        }

        stage('打包后端部分') {
            steps {
                // 下载git仓库的代码
                git credentialsId: 'gitee-repo', url: 'https://gitee.com/danqingzhao/lap.git'

                // 删除 .git 等文件
                sh 'rm -rf .git .gitignore'
                
                // 删除历史静态文件
                sh 'rm -rf html/static html/index.html html/favicon.ico'

                // 创建 logs 目录结构
                sh 'mkdir -p logs .tmp'

                // 加密源码
                sh '/usr/local/lib/python3.11/bin/pyarmor gen core apps main.py'

                // 将源码替换成密文
                sh '''
                    rm -rf core apps main.py
                    mv dist/* .
                    rm -rf dist
                '''
            }
        }

        stage('打包前端部分') {
            steps {
                dir("vue") {
                    // 下载git仓库的代码
                    git credentialsId: 'gitlab-repo', url: 'http://192.168.10.12/jxhu/lap-vue.git'

                    // 编译vue项目
                    nodejs('nodejs-16') {
                        sh 'yarn install'
                        sh 'yarn build:prod'   
                    }
                }
            }
        }

        stage('打包broker') {
            steps {
                dir("broker") {
                    // 下载git仓库的代码
                    sh 'git clone git@192.168.10.12:wyxu/broker3.git --recursive -b v4'

                    // 编译broker
                    sh 'cd broker3 && sh build.sh'

                    // 打包环境
                    sh 'cd broker3 && sh release.sh'
                }
            }
        }

        stage('归档构建产物') {
            steps {
                script {
                    // 将vue的打包文件移动到html文件夹下
                    sh 'mv vue/dist/* html/'

                    // 清理vue项目
                    sh 'rm -rf vue*'

                    // 将broker的打包文件移动到apps/lap/ 目录下
                    sh 'mv broker/broker3/dist apps/lap/'

                    // 删除broker项目
                    sh 'rm -rf broker*'

                    def tempDir = "${env.jenkinTMP}/lap-${env.version}"

                    // 创建临时目录并复制内容
                    sh "mkdir -p ${tempDir}"
                    sh "cp -r * ${tempDir}/"

                    // 打包临时目录内容
                    sh "cd ${env.jenkinTMP} && tar -czf ${env.tarFileName} lap-${env.version}"

                    // 将生成的 tar.gz 移回当前目录
                    sh "mv ${env.jenkinTMP}/${env.tarFileName} ."
                }
            }
        }
        
        stage('上传成品库') {
            steps {
                script {
                    def parts = "${version}".tokenize('.')
                    def remoteDir = "${env.DEPLOY_PATH}/${env.JOB_NAME}/${parts[0..2].join(".")}"
                    def remoteFullPath = "${remoteDir}/${env.tarFileName}"
                    sh """
                        ssh root@192.168.11.54 "mkdir -p ${remoteDir}"
                        scp ${tarFileName} root@192.168.11.54:${remoteFullPath}
                    """
                }
            }
        }
    }


    post {
        always {
            // script {
            //     sh "rm -rf ${env.jenkinTMP}"
            //     sh "rm -rf *"
            // }
            echo "清理打包缓存"
        }
        success {
            echo "构建成功。"
        }
        failure {
            echo "构建失败，请检查日志。"
        }
    }
}