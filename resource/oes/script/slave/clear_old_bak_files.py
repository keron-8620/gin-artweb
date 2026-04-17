#!/usr/local/lib/python3.11/bin/python3.11
import os
import optparse
import shutil
from datetime import datetime, date


def main():
    parser = optparse.OptionParser()
    parser.add_option(
        "-p",
        "--path",
        type="string",
    )
    parser.add_option(
        "-k",
        "--key_word",
        default='txlog',
        type="string",
    )
    parser.add_option(
        "-d",
        "--clear_data",
        type="int",
    )
    (options, _) = parser.parse_args()
    path = options.path.strip()
    key_word = options.key_word.strip()
    clear_data = options.clear_data
    today = date.today()
    for log_data in os.listdir(path):
        d = datetime.strptime(log_data, '%Y%m%d').date()
        if (today - d).days > clear_data:
            log_dir = os.path.join(path, log_data)
            for filename in os.listdir(log_dir):
                if key_word.lower() in filename.lower():
                    filepath = os.path.join(log_dir, filename)
                    if os.path.isfile(filepath):
                        os.remove(filepath)
                    elif os.path.isdir(filepath):
                        shutil.rmtree(filepath)


if __name__ == '__main__':
    main()
