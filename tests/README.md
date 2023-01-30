These scripts run some integration tests with both client and agent processes.  They will start 1 client and 1 agent process and the parameters can vary the intensity of the clients.

Currently, `run_benchmark.py` can be run and has a few command line arguments you can vary.

    python3 run_benchmark.py example example.out

To run multiple benchmarks, simply run `run_benchmarks.py`.  The set of parameters that are used are hard-coded in `run_benchmarks.py`; edit it directly if you want to change them.
