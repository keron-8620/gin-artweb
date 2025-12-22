import os


if __name__ == "__main__":
    ANSIBLE_LOG_PATH = os.getenv("SCRIPT_LOG_PATH")
    ANSIBLE_BASE_DIR = os.getenv("SCRIPT_BASE_DIR")
    print(f"ANSIBLE_LOG_PATH: {ANSIBLE_LOG_PATH}")
    print(f"ANSIBLE_BASE_DIR: {ANSIBLE_BASE_DIR}")
