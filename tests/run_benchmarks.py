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
threads = [1, 2, 4, 6, 8]
pool_size = 25000
buffer_sizes = [128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072]
payload_sizes = [1000]
tracepoints = 128

def make_cmd(args, threads, payload_size, buffer_size, name):
    return [str(v) for v in [
        "python3", "run_benchmark.py",
        "--threads", threads,
        "--duration", duration,
        "--buffer_size", buffer_size,
        "--buffer_count", pool_size,
        "--payload_size", payload_size,
        "--tracepoints", tracepoints,
        "--silent",
        shmname,
        "%s/%s.out" % (args.outdir, name)
    ]]

def run_experiments(args):
    print("running: %s" % str(threads))
    for buffer_size in buffer_sizes:
        for thread in threads:
            for payload_size in payload_sizes:
                name = "%dthreads_%dpayload_%dbufsize" % (thread, payload_size, buffer_size)
                cmd = make_cmd(args, thread, payload_size, buffer_size, name)
                runner = subprocess.Popen(cmd)
                runner.wait()


def process_output(outdir):
    df = None
    for thread in threads:
        for payload_size in payload_sizes:
            for buffer_size in buffer_sizes:
                filename = "%s/%dthreads_%dpayload_%dbufsize.out" % (outdir, thread, payload_size, buffer_size)
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
                    row["payload_size"] = payload_size
                    row["buffer_size"] = buffer_size
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
    df["total_released"] = df["null_released"]+df["pool_released"]
    df["released_bytes"] = df["total_released"] * df["buffer_size"]
    df["goodput_bufs"] = df["pool_released"] * (df["traces"] - df["invalidtraces"]) / df["traces"]
    df["goodput_bytes"] = df["released_bytes"] * (df["traces"] - df["invalidtraces"]) / df["traces"]
    means = df.groupby(["thread", "payload_size", "buffer_size"])[["traces", "invalidtraces", "tracepoints", "bytes", "null_released", "pool_released", "total_released", "released_bytes", "goodput_bufs", "goodput_bytes"]].mean()


    # means = df.groupby("thread")[["begin", "tracepoint", "end"]].mean()
    means.to_csv("scatter2.out")