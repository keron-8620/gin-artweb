#!/usr/local/lib/python3.11/bin/python3.11
import os
import sys


def main():
    path = sys.argv[1]
    if os.path.isdir(path):
        for root, _, files in os.walk(path):
            for name in files:
                os.remove(os.path.join(root, name))
        return
    elif os.path.isfile(path):
        os.remove(path)


if __name__ == '__main__':
    main()
