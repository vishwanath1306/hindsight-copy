# Building

Hindsight comprises the following main pieces:

* A client library written in C, located in the `client` directory
* A local agent written in Go, located in the `agent` directory
* A central log collector written in Go, located in the `agent` directory

## Building client

To build, run tests, and run the go agents:
```
cd client
make
```
This will build several binaries (located in the `bin` folder) and a lib (located in the `lib` folder)

To use the Hindsight client library in your application:
```
sudo make install
```

## Building agent

To check if the go application runs correctly, you can run:
```
cd agent
go run cmd/agent2/main.go
```
By default you would see:
```
config file loaded, addr = 127.0.0.1 port = 5050
running server
/dev/shm/__pool does not exist, waiting...
```
