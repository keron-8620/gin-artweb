#!/usr/local/lib/python3.11/bin/python3.11
import os
import tarfile
import time
from multiprocessing import Process, Queue
import optparse


def _process_archive(queue, backup_path):
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


def multiprocess_archive(src_path, backup_path):
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
    parser = optparse.OptionParser()
    parser.add_option(
        "-s",
        "--src_path",
        type="string",
    )
    parser.add_option(
        "-b",
        "--backup_path",
        type="string",
    )
    (options, _) = parser.parse_args()
    src_path = options.src_path.strip()
    backup_path = options.backup_path.strip()
    multiprocess_archive(src_path, backup_path)


if __name__ == '__main__':
    main()
