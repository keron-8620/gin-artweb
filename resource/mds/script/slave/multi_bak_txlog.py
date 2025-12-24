import os
from pathlib import Path
import time
import tarfile
import argparse
from multiprocessing import Process, Queue


def _process_archive(queue: Queue, backup_path: str):
    while True:
        try:
            filepath = queue.get(block=True, timeout=0.1)
        except Exception:
            break
        tar_name = os.path.basename(filepath)
        tar_file = os.path.join(backup_path, '{}-{}'.format(tar_name, time.strftime('%Y%m%d-%H%M.tar.gz')))
        with tarfile.open(tar_file, 'w:gz') as tar_file:
            tar_file.add(filepath, arcname=tar_name)
            time.sleep(0.1)


def multiprocess_archive(src_path: str, backup_path: str):
    file_queue = Queue()
    process_list = []
    for filename in os.listdir(src_path):
        file_path = os.path.join(src_path, filename)
        if os.path.isfile(file_path):
            file_queue.put(file_path)
    for _ in range(10):
        p = Process(target=_process_archive, args=(file_queue, backup_path))
        p.daemon = True
        p.start()
        time.sleep(0.1)
        process_list.append(p)
    for p in process_list:
        p.join()


def main():
    parser = argparse.ArgumentParser(description="这是一个用于多进程备份文件夹的脚本")
    parser.add_argument(
        "-s",
        "--src_path",
        type=str, 
        required=True,
        help="请输入源路径"
    )
    parser.add_argument(
        "-b",
        "--backup_path",
        type=str, 
        required=True,
        help="请输入备份路径"
    )
    options = parser.parse_args()
    src_path = options.src_path
    backup_path = options.backup_path
    if not os.path.exists(src_path):
        raise FileNotFoundError(f"源路径不存在: {src_path}")
    if not os.path.exists(backup_path):
        os.makedirs(backup_path, exist_ok=True)
    multiprocess_archive(src_path, backup_path)


if __name__ == '__main__':
    main()
