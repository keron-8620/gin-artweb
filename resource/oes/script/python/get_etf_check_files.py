#!/usr/bin/env python3
from typing import Optional, List, Dict
from pathlib import Path
import time
import argparse
from collections import namedtuple
from enum import Enum
import traceback

import yaml

BASE_DIR = Path(__file__).parents[4].absolute()
STORAGE_DIR = BASE_DIR.joinpath("storage")
OES_DIR = STORAGE_DIR.joinpath("oes")
ETF_HEADER = ["etf_id", "security_id", "market", "tradable"]


class Market(Enum):
    sh_mkt = 1
    sz_mkt = 2


def parse_csv(path: Path, headers: list) -> List:
    csv_list = []

    data_namedtuple = namedtuple("CsvLine", headers)
    with open(path) as f:
        for line in f:
            if line.startswith("#") or line.strip() == "":
                continue

            line_data = [i.strip() for i in line.split("|")]
            csv_list.append(data_namedtuple(**dict(zip(headers, line_data))))
        return csv_list


def get_etc_check_files(path: Path):
    sh_files = []
    sz_files = []
    try:
        lines = parse_csv(path, headers=ETF_HEADER)
        for line in lines:
            if str(line.tradable) == str(0):
                continue
            if int(line.market) == Market.sh_mkt.value:
                sh_files.append(
                    f"({line.security_id}{{{{ curr_date[4:] }}}}.(?i)ETF|{line.security_id}{{{{ curr_date[4:] }}}}2.(?i)ETF|ssepcf_{line.security_id}_{{{{ curr_date }}}}.xml)"
                )
            elif int(line.market) == Market.sz_mkt.value:
                # pcf_159942_20180201.xml
                sz_files.append(f"pcf_{line.security_id}_{{{{ curr_date }}}}.xml")
    except Exception:
        print(traceback.format_exc())
        sh_files.append("no_such_etf_file")
        sz_files.append("no_such_etf_file")
    return sh_files, sz_files


def create_sse_var_file(colony_num: str, automatic: Dict, mon_etf_path: Path, counter_etf_path: Path):
    sse_etf_check_mon_files = automatic["sse_etf_check_mon_files"] or []
    sse_etf_check_counter_files = automatic["sse_etf_check_counter_files"]
    all_etf_files = automatic["sse_etf_check_files"] or []
    if sse_etf_check_mon_files:
        mon_etf_files, _ = get_etc_check_files(mon_etf_path)
        all_etf_files += mon_etf_files
    if sse_etf_check_counter_files:
        counter_etf_files, _ = get_etc_check_files(counter_etf_path)
        all_etf_files += counter_etf_files
    tmp_path = OES_DIR.joinpath(".tmp", colony_num, "sse_etf.yaml")
    if not tmp_path.parent.exists():
        tmp_path.parent.mkdir(parents=True, exist_ok=True)
    with open(tmp_path, "w") as f:
        yaml.dump({"etf_check_files": list(set(all_etf_files))}, f)


def create_szse_var_file(colony_num: str, automatic: Dict, mon_etf_path: Path, counter_etf_path: Path):
    all_etf_files = automatic["szse_etf_check_files"] or []
    szse_etf_check_mon_files = automatic["szse_etf_check_mon_files"]
    szse_etf_check_counter_files = automatic["sse_etf_check_counter_files"]
    if szse_etf_check_mon_files:
        _, mon_etf_files = get_etc_check_files(mon_etf_path)
        all_etf_files += mon_etf_files
    if szse_etf_check_counter_files:
        _, counter_etf_files = get_etc_check_files(counter_etf_path)
        all_etf_files += counter_etf_files
    tmp_path = OES_DIR.joinpath(".tmp", colony_num, "szse_etf.yaml")
    if not tmp_path.parent.exists():
        tmp_path.parent.mkdir(parents=True, exist_ok=True)
    with open(tmp_path, "w") as f:
        yaml.dump({"etf_check_files": list(set(all_etf_files))}, f)


def main(task: str, colony_num: str, curr_date: Optional[str] = None):
    automatic_path = OES_DIR.joinpath("config", colony_num, "all", "automatic.yaml")
    if not automatic_path.exists():
        raise FileNotFoundError(f"no such file: {automatic_path}")
    with open(automatic_path, "r") as f:
        automatic = yaml.safe_load(f)
    mon_etf_path = OES_DIR.joinpath("mon", colony_num, "EtfTradeList.csv")
    counter_etf_path = OES_DIR.joinpath("counter", colony_num, "data", "broker", f"EtfTradeList{str(curr_date)[4:]}.csv")
    if task == "sse":
        create_sse_var_file(colony_num, automatic, mon_etf_path, counter_etf_path)
    elif task == "szse":
        create_szse_var_file(colony_num, automatic, mon_etf_path, counter_etf_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="本脚本用于执行oes的playbook")
    parser.add_argument(
        "--task", 
        type=str, 
        help="请输入任务类型(sse或szse)",
        required=True
    )
    parser.add_argument(
        "--colony_num", 
        type=str, 
        help="请输入oes集群编号",
        required=True
    )
    parser.add_argument(
        "--curr_date", 
        type=str, 
        default="",
        help="请输入日期",
    )
    options = parser.parse_args()
    task = options.task.strip()
    colony_num = options.colony_num.strip()
    curr_date = options.curr_date.strip()
    if not curr_date:
        curr_date = time.strftime("%Y%m%d")
    main(task, colony_num, curr_date)
