# Zeus Overview

Zeus is composed of three components:

1. [The Master Process](../go/zeusmaster). This is written in Go, and coordinates all the other processes. It connects Clients to Workers and handles reloading when files change.

2. [Clients](../go/zeusclient). The Client is also written in Go. It sends a command to the Master, and has its streams wired up to a Command process, to make it appear to be running locally.

3. [Workers/Commands](../rubygem). These are the target application. A small shim, written in the target language, manages the communication between the application and the Master process, and boots the application in phases. Though the Master and Client are completely language-agnostic, currently ruby is the only language for which a Worker shim exists.

If you've read Tony Hoare's (or C.A.R. Hoare's) "Communicating Sequential Processes", [`csp.pdf`](http://www.usingcsp.com/cspbook.pdf) might be a bit helpful in addition to this document. I haven't studied the math enough for it to be fully correct, but it gets some of the point across.

See: [`terminology.md`](terminology.md)

## Master Process

### Logical Modules

1. Config

2. ClientHandler

3. FileMonitor

4. WorkerMonitor

![arch.png](arch.png)

The Master process revolves around the [`ProcessTree`](../go/processtree/processtree.go) -- the core data structure that maintains most of the state of the application. Each module performs most of its communication with other modules through interactions with the Tree.

### 1. Config

This component reads the configuration file on initialization, and constructs the initial `ProcessTree` for the rest of the application to use.

* [`config.go`](../go/config/config.go)
* [`zeus.json`](../examples/zeus.json)

### 2. ClientHandler

The `ClientHandler` listens on a socket for incoming requests from Client processes, and negotiates connections to running Worker processes. It is responsible for interactions with the client for its entire life-cycle.

* [`clienthandler.go`](../go/clienthandler/clienthandler.go)

### 3. FileMonitor

The `FileMonitor`'s job is to restart workers when one of their dependencies has changed. Workers are expected to report back with a list of files they have loaded. The `FileMonitor` listens for these messages and registers them with an external process that watches the filesystem for changes. When the external process reports a change, the `FileMonitor` restarts any workers that have loaded that file.

* [`filemonitor.go`](../go/filemonitor/filemonitor.go)
* [`fsevents/main.m`](../ext/fsevents/main.m)

### 4. WorkerMonitor

This component is responsible for communication with the target-language shim to manage booting and forking of application phase workers. It constantly attempts to keep all workers booted, restarting them when they are killed or die.

* [`workermonitor.go`](../go/processtree/workermonitor.go)
* [`workernode.go`](../go/processtree/workernode.go)
* [`master_worker_handshake.md`](master_worker_handshake.md)

## Client Process

The client process is mostly about terminal configuration. It opens a PTY, sets it to raw mode (so that 'fancy' commands behave as if they were running locally), and passes the worker side of the PTY to the Master process.

The client then sets up handlers to write STDIN to the PTY, and write the PTY's output to STDOUT. STDIN is scanned for certain escape codes (`^C`, `^\`, and `^Z`), which are sent as signals to the remote process to mimic the behaviour of a local process.

A handler is set up for SIGWINCH, again to forward it to the remote process, and keep both window sizes in sync.

When the remote process exits, it reports its exit status, which the client process then exits with.

* [`zeusclient.go`](../go/zeusclient/zeusclient.go)
* [`client_master_handshake.md`](client_master_handshake.md)

## Worker/Command Processes

The Worker processes boot the actual application, and run commands. See [`master_worker_handshake.md`](master_worker_handshake.md), and the ruby implementation in the `rubygem` directory.

* [`zeus.rb`](../rubygem/lib/zeus.rb)
* [`zeus/rails.rb`](../rubygem/lib/zeus/rails.rb)

## Contributing to Zeus

See the handy contribution guide at [`docs/contributing.md`](../contributing.md).

