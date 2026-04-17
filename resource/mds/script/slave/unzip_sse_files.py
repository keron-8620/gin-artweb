import os
import sys
import zipfile


def main():
    path = sys.argv[1]
    if os.path.exists(path) and os.path.isdir(path):
        for filename in os.listdir(path):
            if filename.endswith('zip'):
                zip_file = zipfile.ZipFile(os.path.join(path, filename), 'r')
                for file in zip_file.namelist():
                    zip_file.extract(file, path)
                zip_file.close()


if __name__ == '__main__':
    main()
