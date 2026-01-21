#!/usr/bin/env python3
from typing import List, Dict
import os
import sys
import time
from pathlib import Path
import argparse

import yaml
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
OES_DIR = STORAGE_DIR.joinpath("oes")
RESOURCE_DIR = BASE_DIR.joinpath("resource")
SCRIPT_DIR = RESOURCE_DIR.joinpath("oes", "script")
PLAYBOOK_DIR = RESOURCE_DIR.joinpath("oes", "playbook")


def get_curr_date() -> str:
    return time.strftime('%Y%m%d', time.localtime())


def next_trd_date(trd_dates: Dict[str, List[int]], date: int, the_year: int) -> str:
    if not date:
        raise AssertionError('日期不能为空')
    trdDateList = trd_dates.get(f'trd_date_{the_year}_list', [])
    if not trdDateList:
        raise AssertionError(f'交易日历缺少{the_year}年的交易日列表')
    if date >= trdDateList[-1]:
        new_year = the_year + 1
        new_trdDateList = trd_dates.get(f'trd_date_{new_year}_list')
        if not new_trdDateList:
            raise AssertionError(f'交易日历缺少{new_year}年的交易日列表')
        return str(new_trdDateList[0])
    if date in trdDateList:
        index = trdDateList.index(date)
        next_date = trdDateList[index + 1]
        return str(next_date)
    else:
        new_date = date + 1
        if new_date in trdDateList:
            return str(new_date)
        return next_trd_date(trd_dates, new_date, the_year)


def pre_trd_date(trd_dates: Dict[str, List[int]], date: int, the_year: int) -> str:
    if not date:
        raise AssertionError('日期不能为空')
    trdDateList = trd_dates.get(f'trd_date_{the_year}_list', [])
    if not trdDateList:
        raise AssertionError(f'交易日历缺少{the_year}年的交易日列表')
    if date <= trdDateList[0]:
        last_year = the_year -1
        last_trdDateList = trd_dates.get(f'trd_date_{last_year}_list')
        if not last_trdDateList:
            raise AssertionError(f'交易日历缺少{last_year}年的交易日列表')
        return str(last_trdDateList[-1])
    if date in trdDateList:
        index = trdDateList.index(date)
        return str(trdDateList[index - 1])
    else:
        last_date = date - 1
        if last_date in trdDateList:
            return str(last_date)
        return pre_trd_date(trd_dates, last_date, the_year)


def load_mon_conf(mon_id: int) -> Dict:
    """
    加载mon配置

    :param mon_id: mon主机的id
    :return: mon的配置
    """
    mon_path = MON_DIR.joinpath("config", str(mon_id), "mon.yaml")
    if not mon_path.exists():
        raise FileNotFoundError(f"没有这个文件: {mon_path}")
    with open(mon_path, "r") as f:
        mon_vars = yaml.safe_load(f)
    mon_host_path = HOST_CONF_DIR.joinpath(f"host_{mon_vars['host_id']}.yaml")
    if not mon_host_path.exists():
        raise FileNotFoundError(f"没有这个文件: {mon_host_path}")
    with open(mon_host_path, "r") as f:
        mon_host = yaml.safe_load(f)
    return {**mon_vars, **mon_host}



def init_vars(config_path: Path, extravars: str = ""):
    """
    初始化vars配置

    :param config_all_path: 公共配置文件路径
    :param extravars: 额外变量
    :return: vars配置
    """
    colony_path = config_path.joinpath("all", "colony.yaml")
    if not colony_path.exists():
        raise FileNotFoundError(f"没有这个文件: {colony_path}")
    with open(colony_path, "r") as f:
        vars = yaml.safe_load(f)
    vars["mon_host"] = load_mon_conf(vars["mon_node_id"])
    if extravars:
        for item in extravars.split(";"):
            if "=" in item:
                key, value = item.split("=", 1)
                vars[key.strip()] = value.strip()
    trd_data_path = BASE_DIR.joinpath("config", "TrdDateList.yaml")
    if not trd_data_path.exists():
        raise FileNotFoundError(f"缺少交易日历文件: {trd_data_path}")
    with open(trd_data_path, "r") as f:
        trd_dates = yaml.safe_load(f)
    if "curr_date" not in vars:
        vars["curr_date"] = get_curr_date()
    curr_date_int = int(vars["curr_date"])
    curr_year = int(vars["curr_date"][:4])
    if "next_trd_date" not in vars:
        vars["next_trd_date"] = next_trd_date(trd_dates, curr_date_int, curr_year)
    if "pre_trd_date" not in vars:
        vars["pre_trd_date"] = pre_trd_date(trd_dates, curr_date_int, curr_year)
    vars["local_path_script_home"] = str(SCRIPT_DIR)
    vars["local_path_playbook_home"] = str(PLAYBOOK_DIR)
    vars["local_path_oes_home"] = str(OES_DIR)
    vars["local_python_interpreter"] = sys.executable
    return vars


def init_hosts(colony_num: str, conf_path: Path) -> Dict:
    """
    初始化hosts配置

    :param colony_num: oes集群编号
    :return: hosts配置
    """
    oes_cluster_hosts = {}
    for path in conf_path.iterdir():
        if path.is_dir() and path.name in ("host_01", "host_02", "host_03"):
            with open(path.joinpath("node.yaml"), "r") as f:
                node_data = yaml.safe_load(f)
            host_path = HOST_CONF_DIR.joinpath(f"host_{node_data['host_id']}.yaml")
            if not host_path.exists():
                raise FileNotFoundError(f"没有这个文件: {host_path}")
            with open(host_path, "r") as f:
                host_data = yaml.safe_load(f)
            oes_cluster_hosts[f"oes_{colony_num}_{node_data['node_role']}"] = {**node_data, **host_data}
    return oes_cluster_hosts


def main(options):
    playbook_path = PLAYBOOK_DIR.joinpath(options.playbook_path)
    if not (playbook_path.exists() and playbook_path.is_file()):
        raise FileNotFoundError(f"没有这个playbook文件: {playbook_path}")
    colony_num = options.colony_num
    if not colony_num:
        raise ValueError("参数colony_num是必填项")
    config_path = OES_DIR.joinpath("config", colony_num)
    if not config_path.exists():
        raise FileNotFoundError(f"没有这个目录: {config_path}")
    vars = init_vars(config_path, options.extravars)
    vars["colony_num"] = colony_num
    hosts = init_hosts(colony_num, config_path)
    envvars = {"ANSIBLE_LOG_PATH": ANSIBLE_LOG_PATH} if options.enable_ansible_log else {"ANSIBLE_NOCOLOR": "1"}
    return ansible_runner.run(
        inventory={"all": {"hosts": hosts, "vars": vars}},
        playbook=str(playbook_path),
        envvars=envvars,
        verbosity=options.verbosity,
    )
    

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="本脚本用于执行oes相关的playbook任务")
    parser.add_argument(
        "--colony_num", 
        type=str, 
        help="请输入oes集群编号",
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
    options = parser.parse_args()
    result = main(options)
    sys.exit(0) if result.status == "successful" else sys.exit(1)
