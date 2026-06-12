#!/usr/bin/env python3
import time
import os
import argparse
import subprocess
from datetime import datetime

from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler


def exec_rsync(ssh_host, ssh_port, ssh_user, src_dir, dest_dir):
    cmd = [
        "rsync",
        "-e", f"ssh -p {ssh_port}",
        "-apz", "--delete",
        f"{src_dir}/",
        f"{ssh_user}@{ssh_host}:{dest_dir}/"
    ]
    print(" ".join(cmd))
    try:
        result = subprocess.run(cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, universal_newlines=True)
        return True, result.stdout
    except subprocess.CalledProcessError as e:
        print(f"同步失败: {e.stderr}")
        return False, e.stderr
    except Exception as e:
        print(f"执行异常: {str(e)}")
        return False, str(e)


class FileEventHandler(FileSystemEventHandler):
    def __init__(self, ssh_host, ssh_port, ssh_user, src_dir, dst_dir, sync_interval= 2):
        super().__init__()
        self.ssh_host = ssh_host
        self.ssh_port = ssh_port
        self.ssh_user = ssh_user
        self.src_dir = src_dir
        self.dst_dir = dst_dir
        self.sync_interval = sync_interval
        self.last_sync = datetime.now()

    def on_modified(self, event):
        self._sync_with_debounce()

    def on_created(self, event):
        self._sync_with_debounce()

    def on_deleted(self, event):
        self._sync_with_debounce()

    def _sync_with_debounce(self):
        now = datetime.now()
        if (now - self.last_sync).seconds < self.sync_interval:
            return
        self.last_sync = now
        exec_rsync(self.ssh_host, self.ssh_port, self.ssh_user, self.src_dir, self.dst_dir)


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
