# Testing

## Unit Tests

After [building](building.md) you can run the single-process unit tests:
```
bin/queue_test
bin/buffer_test
```
You should see output such as:
```
Testing queue implementation
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_queue_simple (0x7f7a65016000)
test_queue_simple passedd
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_queue_nonblocking (0x7f7a65015000)
test_queue_nonblocking passed
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_queue_blocking (0x7f7a65014000)
test_queue_blocking passed
Created queue capacity=1 element_size=4 element_total_size=8 at /dev/shm/test_tiny_queue (0x7f7a65013000)
test_tiny_queue passed
Created queue capacity=10 element_size=80 element_total_size=84 at /dev/shm/test_queue_struct (0x7f7a65012000)
test_queue_struct passed
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_queue_blocking_multithread (0x7f7a65011000)
  test_queue_blocking_multithread: thread 0 start
  test_queue_blocking_multithread: thread 7 start
  test_queue_blocking_multithread: thread 3 start
  test_queue_blocking_multithread: thread 2 start
  test_queue_blocking_multithread: thread 4 start
  test_queue_blocking_multithread: thread 5 start
  test_queue_blocking_multithread: thread 6 start
  test_queue_blocking_multithread: thread 9 start
  test_queue_blocking_multithread: thread 8 start
  test_queue_blocking_multithread: thread 1 start
  test_queue_blocking_multithread: thread 4 exit
  test_queue_blocking_multithread: thread 1 exit
  test_queue_blocking_multithread: thread 2 exit
  test_queue_blocking_multithread: thread 3 exit
  test_queue_blocking_multithread: thread 0 exit
  test_queue_blocking_multithread: thread 8 exit
  test_queue_blocking_multithread: thread 7 exit
  test_queue_blocking_multithread: thread 6 exit
  test_queue_blocking_multithread: thread 9 exit
  test_queue_blocking_multithread: thread 5 exit
test_queue_blocking_multithread passed
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_queue_nonblocking_multi (0x7f7a65010000)
test_queue_nonblocking_multi passed
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_queue_nonblocking_multi (0x7f7a6500f000)
test_queue_put_nonblocking_multi passed
```
```
Testing buffer!
test_buffer_simple passed
test_buffer_write passed
Created buffer pool, capacity=10 buffer_size=100 at /dev/shm/test_bufmanager__pool
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_bufmanager__available_queue (0x7f8f730c1000)
Created queue capacity=10 element_size=16 element_total_size=20 at /dev/shm/test_bufmanager__complete_queue (0x7f8f730c0000)
test_bufmanager passed
Created buffer pool, capacity=10 buffer_size=100 at /dev/shm/test_tracestate__pool
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_tracestate__available_queue (0x7f8f730be000)
Created queue capacity=10 element_size=16 element_total_size=20 at /dev/shm/test_tracestate__complete_queue (0x7f8f730bd000)
test_tracestate passed
Created buffer pool, capacity=10 buffer_size=100 at /dev/shm/test_tracestate_nullbuffer__pool
Created queue capacity=10 element_size=4 element_total_size=8 at /dev/shm/test_tracestate_nullbuffer__available_queue (0x7f8f730bb000)
Created queue capacity=10 element_size=16 element_total_size=20 at /dev/shm/test_tracestate_nullbuffer__complete_queue (0x7f8f730ba000)
test_tracestate_nullbuffer passed
```

You can run golang unit tests by navigating to the appropriate folder and running:
```
go test *.go -v
```

## Integration Tests

There are several integration tests that use different combinations of C and Go agents

### Integration Test 1: C client, C agent

These integration tests check if the shared memory queues are working and display simple throughput numbers.

The tests require two terminals.  Note: terminal 1 must be run before terminal 2.

**Terminal 1:**
```
bin/hindsight_test client
```
You should see output that looks like:
```
Running as client
Hindsight Config:
  Buffer pool cap=10000 buf_length=4000
  Service addr=
  Queue sizes breadcrumbs_cap=10000 triggers_cap=10000
Created buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Created queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7f2b440ad000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7f2b4407c000)
Created queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7f2b417e8000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7f2b417b7000)
Beginning client trace
Beginning client loop
Tracepoints 22739314 - Pool: 0 0 - NULL 5912221 5912220
Tracepoints 16846041 - Pool: 1048217 1048216 - NULL 3331752 3331753
Tracepoints 9267321 - Pool: 2409503 2409503 - NULL 0 0
Tracepoints 9310356 - Pool: 2420692 2420692 - NULL 0 0
```

**Terminal 2:**
```
bin/hindsight_test
```
You should see output that looks like:
```
Running as agent
Loaded existing buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Loaded existing queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7f3e7f905000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7f3e7f8d4000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7f3e7d08e000)
Loaded existing queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7f3e7d00f000)
Inited existing bufmanager hs_integration_test
Resetting buffers: draining any existing buffers...
Resetting buffers: drained 0 available and 0 complete.
Initialize buffers: making 10000 buffers available...
Initialize buffers: done
Queue states:
  Available occupancy=8764 remaining=1236 head=1236 tail=10000
  Complete occupancy=1249 remaining=8751 head=0 tail=1249
Throughput: 2430773
Throughput: 2443662
Throughput: 2444888
```

***Note:*** If the second terminal is stuck and shows no throughput, make sure you ran Terminal 1 **before** running terminal 2.  

### Integration Test 2: C Client, Go Agent

These integration tests check if the shared memory queues are working between C and Go, and display simple throughput numbers.  
The expected output should look very similar to the first integration test, except now the agent runs in Go.

The tests require two terminals.  Note: terminal 1 must be run before terminal 2.

**Terminal 1:**
```
cd client
bin/hindsight_test client
```
You should see output that looks like:
```
Running as client
Hindsight Config:
  Buffer pool cap=10000 buf_length=4000
  Service addr=
  Queue sizes breadcrumbs_cap=10000 triggers_cap=10000
Created buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Created queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7fecc88d3000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7fecc88a2000)
Created queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7fecc600e000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7fecc5fdd000)
Beginning client trace
Beginning client loop
Tracepoints 22695729 - Pool: 0 0 - NULL 5900889 5900888
Tracepoints 20517919 - Pool: 439310 439309 - NULL 4895348 4895349
Tracepoints 11820178 - Pool: 3073245 3073245 - NULL 0 0
```

**Terminal 2:**
```
cd agent
go run cmd/jon/jon.go
```
You should see output that looks like:
```
Hello world!
Loaded existing buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Loaded existing queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7f2b7dd4b000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7f2b7dd1a000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7f2b7dce9000)
Loaded existing queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7f2b7dc6a000)
Inited existing bufmanager hs_integration_test
Resetting buffers: drained 0 available and 0 complete
Initialize buffers: making 10000 buffers available...
Initialize buffers: done
Queue states:
  Available occupancy=8698 remaining=1302 head=1302 tail=10000
  Complete occupancy=1327 remaining=8673 head=0 tail=1327
Throughput: 3051905 Average batch: 5.4478493
Throughput: 3085956 Average batch: 5.9465938
Throughput: 3069194 Average batch: 5.5231023
```
***Note:*** As with Integration test 1, terminal 1 must be executed before terminal 2, otherwise you will see no throughput measurements from terminal 2.


### Integration Test 3: C Client, Go Agent (2)

These integration tests check if the shared memory queues are working between C and Go, using the Go-based channel API, and display simple throughput numbers.  
The expected output should look very similar to the first two integration tests, except now the throughput might be slightly lower.

The tests require two terminals.  Note: terminal 1 must be run before terminal 2.

**Terminal 1:**
```
cd client
bin/hindsight_test client
```
You should see output that looks like:
```
Running as client
Hindsight Config:
  Buffer pool cap=10000 buf_length=4000
  Service addr=
  Queue sizes breadcrumbs_cap=10000 triggers_cap=10000
Created buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Created queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7f96fd6a2000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7f96fd67100$)
Created queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7f96faddd000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7f96fadac000)
Beginning client trace
Beginning client loop
Tracepoints 22261300 - Pool: 0 0 - NULL 5787938 5787937
Tracepoints 22355585 - Pool: 0 0 - NULL 5812450 5812450
Tracepoints 18776592 - Pool: 531333 531332 - NULL 4350580 4350581
Tracepoints 8069977 - Pool: 2081071 2081071 - NULL 17121 17121
Tracepoints 8026973 - Pool: 2071299 2071299 - NULL 15713 15713
```

**Terminal 2:**
```
cd agent
go run cmd/jon/jon3.go
```
You should see output that looks like:
```
Hello world!
Loaded existing buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Loaded existing queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7fcaf4337000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7fcaf4306000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7fcaf42d5000)
Loaded existing queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7fcaf4256000)
Resetting buffers: drained 0 available and 0 complete
Initialize buffers: making 10000 buffers available...
Initialize buffers: done
Queue states:
  Available occupancy=8559 remaining=1441 head=1441 tail=10000
  Complete occupancy=1472 remaining=8528 head=0 tail=1472
shm queue goroutine running
Throughput: 2133839 Average batch: 7.357556
Throughput: 2072235 Average batch: 5.2623467
Throughput: 2047736 Average batch: 5.4422784
```
***Note:*** As with Integration test 1, terminal 1 must be executed before terminal 2, otherwise you will see no throughput measurements from terminal 2.

### Integration Test 4: C Client, Real Go Agent

This integration test runs the real Go agent and displays simple throughput numbers.  It also prints whenever a trace is triggered using Hindsight's `trigger` API.
You will see RPC errors, as the test attempts to connect to an RPC server that isn't running.

**Prerequisites:** For this test, you must `sudo make install` to ensure the Hindsight configs have been installed, otherwise you may see config errors.

The tests require two terminals.  Note: terminal 1 must be run before terminal 2.

**Terminal 1:**
```
cd client
bin/hindsight_test client
```
***Note:*** *You can alternately run `bin/hindsight2_test client` which generates a lower volume of data*

You should see output that looks like:
```
Running as client
Hindsight Config:
  Buffer pool cap=10000 buf_length=4000
  Service addr=
  Queue sizes breadcrumbs_cap=10000 triggers_cap=10000
Created buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Created queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7f5c24653000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7f5c24622000)
Created queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7f5c21d8e000)
Created queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7f5c21d5d000)
Beginning client trace
Beginning client loop
Tracepoints 22777294 - Pool: 0 0 - NULL 5922096 5922095
Tracepoints 22701496 - Pool: 0 0 - NULL 5902388 5902388
Tracepoints 21699167 - Pool: 0 0 - NULL 5641783 5641783
Tracepoints 14874701 - Pool: 1184282 1184281 - NULL 2683140 2683141
Tracepoints 11341556 - Pool: 1686407 1686407 - NULL 1262396 1262396
```
**Terminal 2:**
```
cd agent
go run cmd/agent2/main.go --serv hs_integration_test
```
You should see output that looks like:
```
config file loaded, addr = 127.0.0.1 port = 5050
running server
Loaded existing buffer pool, capacity=10000 buffer_size=4000 at /dev/shm/hs_integration_test__pool
Loaded existing queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/hs_integration_test__available_queue (0x7f82940c0000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__complete_queue (0x7f829408f000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/hs_integration_test__triggers_queue (0x7f829405e000)
Loaded existing queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/hs_integration_test__breadcrumbs_queue (0x7f827d95b000)
Resetting buffers: drained 0 available and 0 complete
Initialize buffers: making 10000 buffers available...
Initialize buffers: done
Queue states:
  Available occupancy=9771 remaining=229 head=229 tail=10000
  Complete occupancy=243 remaining=9757 head=0 tail=243
Go Agent cache capacity 8000
TriggerManager goroutine running
TraceCache goroutine running
shm queue goroutine running
TriggerManager connected to 127.0.0.1:5252
Throughput: 0 Average batch: 100
Reporting trace 800000 with 26 buffers, breadcrumbs: (0: Hello World!)
report rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:5252: connect: connection refused"
Throughput: 1707489 Average batch: 12.39042
Reporting trace 900000 with 0 buffers, breadcrumbs: (0: Hello World!)
report rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:5252: connect: connection refused"
Throughput: 1702457 Average batch: 11.534275
Reporting trace 1000000 with 0 buffers, breadcrumbs: (0: Hello World!)
report rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:5252: connect: connection refused"
Reporting trace 1000000 with 26 buffers, breadcrumbs:
report rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:5252: connect: connection refused"
Throughput: 1672702 Average batch: 11.423772
```
In the above test, every 100,000 traces gets triggered.


### Integration Test 5: Multiple Clients, Multiple Agents, Coordinator and Triggers

This test runs 4 clients, 4 agents, and a coordinator.  Each client has a breadcrumb to the next.  Only one client fires triggers.  The coordinator will collect trace data from all clients.

Expected output: you should see output from the coordinator terminal demonstrating the breadcrumb traversal process.

**Terminal 0 (coordinator)**
Start by clearing dev shm and running the coordinator
```
rm /dev/shm/*
cd agent
go run cmd/coordinator/main.go
```

**Terminal 1A (agent 1):**
```
cd agent
go run cmd/agent2/main.go --serv hs_breadcrumb_test1 -port 5053
```

**Terminal 2A (agent 2):**
```
cd agent
go run cmd/agent2/main.go --serv hs_breadcrumb_test2 -port 5054
```

**Terminal 3A (agent 3):**
```
cd agent
go run cmd/agent2/main.go --serv hs_breadcrumb_test3 -port 5055
```

**Terminal 4A (agent 4):**
```
cd agent
go run cmd/agent2/main.go --serv hs_breadcrumb_test4 -port 5056
```

**Terminal 1C (client 1):**
```
cd client
bin/trigger_benchmark_test hs_breadcrumb_test1 -S 100000 -a 127.0.0.1:5053
```
*The -S flag adds a sleep to the client.  Lower values sleeps less and produces more data*

**Terminal 2C (client 2):**
```
cd client
bin/trigger_benchmark_test hs_breadcrumb_test2 -S 100000 -a 127.0.0.1:5054 -b 127.0.0.1:5053
```
*The -b flag adds a breadcrumb to the specified address; in this case it is to agent1*

**Terminal 3C (client 3):**
```
cd client
bin/trigger_benchmark_test hs_breadcrumb_test3 -S 100000 -a 127.0.0.1:5055 -b 127.0.0.1:5054
```

**Terminal 4C (client 4):**
```
cd client
bin/trigger_benchmark_test hs_breadcrumb_test4 -S 100000 -a 127.0.0.1:5056 -b 127.0.0.1:5055 -p 1
```
*The -p flag adds a trigger with probability 1 that it will fire -- that is, every request*
