# Documentation

## Building and Configuring Hindsight

* [prerequisites.md](prerequisites.md) - before building Hindsight for the first time
* [building.md](building.md) - how to build Hindsight
* [testing.md](testing.md) - running tests and more elaborate experiments
* The respective documentation for [cliends](clients.md), [agent](agent.md), [coordinator](coordinator.md), and [collector](collector.md) outline the command-line configuration of those processes
* [configuration.md](configuration.md) gives an overview of the Hindsight configuration file
* [workload.md](workload.md) gives an overview of a simple Hindsight workload generator
* [telemetry.md](telemetry.md) explains configuration of Agent telemetry files.

## Running Hindsight Processes

* [clients.md](clients.md) - client applications must be instrumented using Hindsight's client API
* [agent.md](agent.md) - for each client process there must be a corresponding agent process running on the same host that is responsible for indexing and managing the locally-generated trace data
* [coordinator.md](coordinator.md) - a single coordinator process that can run somewhere on the network and is responsible for disseminating breadcrumbs and triggers
* [collector.md](collector.md) - a single collector process that can run somewhere on the network and is responsible for receiving trace data.
