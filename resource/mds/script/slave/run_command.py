#!/usr/local/bin/python3
import sys
import subprocess


def Run_Command(command):
    response_dict = dict()
    try:
        result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
    except Exception as e:
        response_dict['status'] = -1
        response_dict['response'] = e
    else:
        response_dict['status'] = result.returncode
        response_dict['response'] = result.stdout.decode('utf-8') + '\n'
        response_dict['response'] += result.stderr.decode('utf-8') + '\n'
    finally:
        return response_dict

    
if __name__ == '__main__':
    cmd = sys.argv[1]
    Run_Command(cmd)
