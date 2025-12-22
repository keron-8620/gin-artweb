#!/usr/local/bin/python3.8
from typing import Dict
import os
import sys
import time
from pathlib import Path
import argparse
import json

import ansible_runner

ANSIBLE_LOG_PATH = os.getenv("JOB_LOG_PATH")
if not ANSIBLE_LOG_PATH:
    raise AssertionError("环境变量没有设置JOB_LOG_PATH")

ANSIBLE_BASE_DIR = os.getenv("JOB_BASE_DIR")
if not ANSIBLE_BASE_DIR:
    raise AssertionError("环境变量没有设置JOB_BASE_DIR")

BASE_DIR = Path(ANSIBLE_BASE_DIR)
STORAGE_DIR = BASE_DIR.joinpath("storage")
HOST_CONF_DIR = STORAGE_DIR.joinpath("host_vars")
MON_DIR = STORAGE_DIR.joinpath("mon")
RESOURCE_DIR = BASE_DIR.joinpath("resource")
SCRIPT_DIR = RESOURCE_DIR.joinpath("mon", "script")
PLAYBOOK_DIR = RESOURCE_DIR.joinpath("mon", "playbook")


def init_vars(mon_host_id: int, extravars: str = ""):
    """
    初始化vars配置

    :param config_all_path: 公共配置文件路径
    :param extravars: 额外变量
    :return: vars配置
    """
    mon_path = MON_DIR.joinpath("config", str(mon_host_id), "mon.json")
    if not mon_path.exists():
        raise FileNotFoundError(f"没有这个文件: {mon_path}")
    with open(mon_path, "r") as f:
        vars = json.load(f)
    if extravars:
        for item in extravars.split(";"):
            if "=" in item:
                key, value = item.split("=", 1)
                vars[key.strip()] = value.strip()
    if "curr_date" not in vars:
        vars["curr_date"] = time.strftime("%Y%m%d")
    vars["local_path_script_home"] = str(SCRIPT_DIR)
    vars["local_path_playbook_home"] = str(PLAYBOOK_DIR)
    vars["local_path_mon_home"] = str(MON_DIR)
    vars["local_python_interpreter"] = sys.executable
    return vars


def init_hosts(host_id: str) -> Dict:
    """
    初始化hosts配置

    :param colony_num: mon集群编号
    :return: hosts配置
    """
    host_path = HOST_CONF_DIR.joinpath(f"host_{host_id}.json")
    if not host_path.exists():
        raise FileNotFoundError(f"没有这个文件: {host_path}")
    with open(host_path, "r") as f:
        return json.load(f)


def main(options):
    playbook_path = PLAYBOOK_DIR.joinpath(options.playbook_path)
    if not playbook_path.exists() or not playbook_path.is_file():
        raise FileNotFoundError(f"没有这个playbook文件: {playbook_path}")
    mon_id = options.mon_id
    if not mon_id:
        raise ValueError("参数mon_host_id是必填项")
    vars = init_vars(mon_id, options.extravars)
    hosts = {f"mon_{mon_id}": init_hosts(vars["host_id"])}
    return ansible_runner.run(
        inventory={"all": {"hosts": hosts, "vars": vars}},
        playbook=str(playbook_path),
        envvars={
            "ANSIBLE_NOCOLOR": "1", 
            "ANSIBLE_LOG_PATH": options.log_path,
        },
        verbosity=options.verbosity,
    )


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="本脚本用于执行mon相关的playbook任务")
    parser.add_argument(
        "--mon_id", 
        type=int, 
        help="请输入mon主机id",
        required=True
    )
    parser.add_argument(
        "--playbook_path", 
        type=str, 
        help="playbook文件的相对路径",
        required=True
    )
    parser.add_argument(
        "--log_path", 
        type=str, 
        help="请输入日志文件路径",
        required=True
    )
    parser.add_argument(
        "--verbosity", 
        type=int, 
        choices=range(0, 5),
        default=0,
        help="请输入输出详细程度(0-4, 0为最少输出, 4为最详细)",
    )
    parser.add_argument(
        "--extravars", 
        type=str, 
        default="",
        help="请输入额外的变量(a=b,c=d)",
    )
    options = parser.parse_args()
    result = main(options)
    sys.exit(0) if result.status == "successful" else sys.exit(1)
