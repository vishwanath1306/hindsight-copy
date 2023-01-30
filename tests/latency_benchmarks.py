import json
import subprocess
import signal
import argparse
import time
import numpy as np
import math
import os
import pandas as pd

parser = argparse.ArgumentParser(description='Run multiple benchmarks')
parser.add_argument("outdir", metavar="OUT", type=str, help="Output directory")

shmname = "multi"
duration = 60 # Duration per run
threads = list(range(1,33))
buffer_size = 32768
payload_size = 4
tracepoints = 1000

def make_cmd(args, threads, name):
    return [str(v) for v in [
        "python3", "run_benchmark.py",
        "--threads", threads,
        "--duration", duration,
        "--buffer_size", buffer_size,
        "--payload_size", payload_size,
        "--tracepoints", tracepoints,
        "--silent",
        shmname,
        "%s/%s.out" % (args.outdir, name)
    ]]

def run_experiments(args):
    print("running: %s" % str(threads))
    for thread in threads:
        cmd = make_cmd(args, thread, "%dthreads" % thread)
        runner = subprocess.Popen(cmd)
        runner.wait()


def process_output(outdir):
    df = None
    for thread in threads:
        filename = "%s/%dthreads.out" % (outdir, thread)
        with open(filename, "r") as f:
            lines = f.readlines()
        headerline = [l.strip() for l in lines if l.startswith("headers:")][0]
        datalines = [l.strip() for l in lines if l.startswith("data:")]
        datalines = datalines[int(len(datalines)/2):len(datalines)-1]
        headers = headerline.split("\t")[1:]
        data = [l.split("\t")[1:] for l in datalines]
        rows = [dict(zip(headers, d)) for d in data]
        for row in rows:
            row["thread"] = thread
        if df is None:
            df = pd.DataFrame(rows)
        else:
            df = df.append(rows)
    df = df.apply(pd.to_numeric)
    return df




if __name__ == '__main__':
    args = parser.parse_args()
    run_experiments(args)
    df = process_output(args.outdir)

    means = df.groupby("thread")[["begin", "tracepoint", "end"]].mean()
    means.to_csv("latencies2.out")