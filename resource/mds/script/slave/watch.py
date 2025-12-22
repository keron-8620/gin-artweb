#!/usr/local/lib/python3.11/bin/python3.11
import time
import os
import argparse
import subprocess

from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler


def exec_rsync(ssh_host, ssh_port, ssh_user, src_dir, dest_dir):
    command = "rsync -e 'ssh -p %s' -apz --delete %s/ %s@%s:%s/" % (
        ssh_port, src_dir, ssh_user, ssh_host, dest_dir
    )
    print(command)
    try:
        subprocess.run(command, shell=True)
    except Exception as e:
        print(e)


class FileEventHandler(FileSystemEventHandler):
    def __init__(self, ssh_host, ssh_port, ssh_user, src_dirs, ssh_path):
        self.ssh_host = ssh_host
        self.ssh_port = ssh_port
        self.ssh_user = ssh_user
        self.ssh_path = ssh_path
        self.src_dirs = src_dirs
        FileSystemEventHandler.__init__(self)

    def on_any_event(self, event):
        for src_dir in self.src_dirs:
            exec_rsync(self.ssh_host, self.ssh_port, self.ssh_user, src_dir, self.ssh_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-s",
        "--ssh_host",
        type=str,
        help="请输入同步机器的ip"
    )
    parser.add_argument(
        "-p",
        "--ssh_port",
        type=int,
        default=22,
        help="请输入同步机器的端口, 默认为22端口"
    )
    parser.add_argument(
        "-u",
        "--ssh_user",
        help="请输入远程连接用户名"
    )
    parser.add_argument(
        "-d",
        "--ssh_path",
        type=str,
        help="请输入远程同步的基础路径"
    )
    parser.add_argument(
        "-l",
        "--dir_list",
        type=str,
        help="请输入需要同步的文件夹列表, 多个用逗号隔开"
    )
    options = parser.parse_args()
    ssh_host = options.ssh_host.strip()
    ssh_port = options.ssh_port
    ssh_user = options.ssh_user.strip()
    ssh_path = options.ssh_path.strip()
    dir_list = options.dir_list.strip()
    local_path = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
    observer = Observer()
    for dir_name in dir_list.split(","):
        dir_name = dir_name.strip()
        if not dir_name:
            continue
        if dir_name.startswith("/"):
            raise Exception("文件夹名称不能使用绝对路径")
        src_dir = os.path.join(local_path, dir_name)
        dst_dir = os.path.join(ssh_path, dir_name)
        exec_rsync(ssh_host, ssh_port, ssh_user, src_dir, dst_dir)
        event_handler = FileEventHandler(ssh_host, ssh_port, ssh_user, src_dir, dst_dir)
        observer.schedule(event_handler, path=src_dir, recursive=True)
    observer.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()
