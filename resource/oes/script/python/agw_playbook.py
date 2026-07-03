#!/usr/bin/env python3
import os
import sys
from datetime import datetime
from pathlib import Path
import argparse
import tempfile
import yaml
import shutil
import uuid

import ansible_runner

BASE_DIR = Path(__file__).resolve().parents[4]
STORAGE_DIR = BASE_DIR.joinpath("storage")
HOST_CONF_DIR = STORAGE_DIR.joinpath("host_vars")
AGW_DIR = STORAGE_DIR.joinpath("agw")
RESOURCE_DIR = BASE_DIR.joinpath("resource")
SCRIPT_DIR = RESOURCE_DIR.joinpath("oes", "script")
PLAYBOOK_DIR = RESOURCE_DIR.joinpath("oes", "playbook")

JOB_LOG_PATH = os.getenv("JOB_LOG_PATH")
if not JOB_LOG_PATH or not os.path.exists(JOB_LOG_PATH):
    JOB_LOG_PATH = STORAGE_DIR.joinpath("logs", datetime.now().strftime("%Y%m%d"), f"{uuid.uuid4()}.log").as_posix()

JOB_RECORD_ID = os.getenv("JOB_RECORD_ID")
if not JOB_RECORD_ID:
    JOB_RECORD_ID = 0


def init_vars(agw_host_id: int, extravars: str = ""):
    """
    初始化vars配置

    :param config_all_path: 公共配置文件路径
    :param extravars: 额外变量
    :return: vars配置
    """
    agw_path = AGW_DIR.joinpath("config", str(agw_host_id), "agw.yaml")
    if not agw_path.exists():
        raise FileNotFoundError(f"没有这个文件: {agw_path}")
    with open(agw_path, "r") as f:
        vars = yaml.safe_load(f)
    if extravars:
        for item in extravars.split(","):
            if "=" in item:
                key, value = item.split("=", 1)
                vars[key.strip()] = value.strip()
    vars["local_path_agw_home"] = str(AGW_DIR)
    vars["local_path_script_home"] = str(SCRIPT_DIR)
    vars["local_path_playbook_home"] = str(PLAYBOOK_DIR)
    vars["local_python_interpreter"] = sys.executable
    return vars


def init_hosts(host_id: str) -> dict:
    """
    初始化hosts配置

    :param colony_num: mon集群编号
    :return: hosts配置
    """
    host_path = HOST_CONF_DIR.joinpath(f"host_{host_id}.yaml")
    if not host_path.exists():
        raise FileNotFoundError(f"没有这个文件: {host_path}")
    with open(host_path, "r") as f:
        return yaml.safe_load(f)


def main(options):
    playbook_path = PLAYBOOK_DIR.joinpath(options.playbook_path)
    if not playbook_path.exists() or not playbook_path.is_file():
        raise FileNotFoundError(f"没有这个playbook文件: {playbook_path}")
    agw_id = options.agw_id
    if not agw_id:
        raise ValueError("参数agw_id是必填项")
    vars = init_vars(agw_id, options.extravars)
    vars["agw_id"] = str(agw_id)
    hosts = {f"agw_{agw_id}": init_hosts(vars["host_id"])}
    envvars = {}
    if options.enable_ansible_log:
        envvars["ANSIBLE_LOG_PATH"] = JOB_LOG_PATH
    if not options.enable_ansible_color:
        envvars["ANSIBLE_NOCOLOR"] = "1"
    tmpdir = tempfile.mkdtemp()
    try:
        return ansible_runner.run(
            inventory={"all": {"hosts": hosts, "vars": vars}},
            playbook=str(playbook_path),
            envvars=envvars,
            verbosity=options.verbosity,
            private_data_dir=tmpdir
        )
    finally:
        shutil.rmtree(tmpdir)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="本脚本用于执行mon相关的playbook任务")
    parser.add_argument(
        "--agw_id", 
        type=int, 
        help="请输入age的id",
        required=True
    )
    parser.add_argument(
        "--playbook_path", 
        type=str, 
        help="playbook文件的相对路径",
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
    parser.add_argument(
        "--enable_ansible_log", 
        type=bool,
        default=False,
        help="是否启用ansible日志",
    )
    parser.add_argument(
        "--enable_ansible_color", 
        type=bool,
        default=False,
        help="是否启用ansible颜色输出",
    )
    options = parser.parse_args()
    result = main(options)        
    sys.exit(0) if result.status == "successful" else sys.exit(1)
