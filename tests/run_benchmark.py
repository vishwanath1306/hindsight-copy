import json
import subprocess
import signal
import argparse
import time
import numpy as np
import math
import os

parser = argparse.ArgumentParser(description='Run a hindsight benchmark')
parser.add_argument("name", metavar="SERVNAME", type=str, help="Service name")
parser.add_argument("output", metavar="OUT", type=str, help="Output filename")
parser.add_argument('-t', "--threads", metavar="NUM", type=int, default="1", help='Number of threads')
parser.add_argument('-s', "--buffer_size", metavar="NUM", type=int, default="4096", help='Buffer size')
parser.add_argument('-c', "--buffer_count", metavar="NUM", type=int, default="25000", help='Buffer Count')
parser.add_argument('-w', "--payload_size", metavar="NUM", type=int, default="1000", help='Payload size')
parser.add_argument('-n', "--tracepoints", metavar="NUM", type=int, default="100", help='Number of tracepoints per trace')
parser.add_argument('-p', "--trigger", metavar="NUM", type=float, default="0", help='Trigger probability')
parser.add_argument('-H', "--headsampling", metavar="NUM", type=float, default="0", help='Head-based sampling probability')
parser.add_argument('-R', "--retroactive", metavar="NUM", type=float, default="1", help='Retroactive tracing probability')
parser.add_argument('-d', "--duration", metavar="NUM", type=int, default="60", help='Experiment duration')
parser.add_argument('-silent', "--silent", action='store_true', help='Prompt before proceeding')


import pathlib
from pathlib import Path
def find_models(basedir):
    paths = list(Path(basedir).rglob("model.clockwork_params"))
    paths = [str(p)[:-17] for p in paths]
    return paths


def reset_shm(args):
    files = ["complete_queue", "triggers_queue", "pool", "breadcrumbs_queue", "available_queue"]
    for file in files:
        cmd = ["rm", "-v", "/dev/shm/%s__%s" % (args.name, file)]
        child = subprocess.Popen(cmd)
        child.wait()
    print("Reset shm")


def make_client_cmd(args):
    cmd = [str(v) for v in [
        "bin/benchmark_test",
        "--threads", args.threads,
        "--buffer_size", args.buffer_size,
        "--buffer_count", args.buffer_count,
        "--payload_size", args.payload_size,
        "--tracepoints", args.tracepoints,
        "--trigger", args.trigger,
        "--duration", args.duration,
        "--headsampling", args.headsampling,
        "--retroactive", args.retroactive,
        args.name
    ]]
    print(" ".join(cmd))
    return cmd

def make_agent_cmd(args):
    cmd = ["go", "run", "cmd/agent2/main.go", "--serv", args.name]
    return cmd


def run(args):
    if not args.silent:
        print("Delete shm files for %s? Press <return> to continue or CTRL-C to abort" % args.name)
        input()

    reset_shm(args)

    # cmds = [make_client_cmd(args), make_agent_cmd(args)]

    client = subprocess.Popen(make_client_cmd(args), stdout=subprocess.PIPE, cwd="../client")
    # agent = subprocess.Popen(make_agent_cmd(args), stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd="../agent")
    agent = subprocess.Popen(make_agent_cmd(args), stdout=subprocess.PIPE, cwd="../agent", preexec_fn=os.setsid)

    try:
        lines = []
        while True:
            line = client.stdout.readline().decode().strip()
            if not line:
                break
            lines.append(line)

        with open(args.output, "w") as f:
            for line in lines:
                f.write(line + "\n")

        # agent.send_signal(signal.SIGTERM)
        # agent.terminate()
        os.killpg(os.getpgid(agent.pid), signal.SIGINT)
        agent.wait()
    except:
        print("Killing agent")
        os.killpg(os.getpgid(agent.pid), signal.SIGINT)
        agent.wait()
        raise

if __name__ == '__main__':
    args = parser.parse_args()
    exit(run(args))
