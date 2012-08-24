# Overview of the Master Process

The Master process all revolves around a single core data structure -- the `ProcessTree`.

The `ProcessTree` represents a tree of all the processes that should exist when
the application is fully booted. Each process is represented by a node. A node
knows about:

* the pid of the currently-running process (if booted);

* the identifier as specified in the config file;

* The action;

* a list of pointers to child nodes;

* A list of pointers to a different kind of node representing a command; and

* A list of features this node requires that its parent does not.

The command nodes contain:

* the command name (identifier);

* A list of aliases; and

* The action.

There are four main components to the Master process software. Each of these components is largely interested in working with the ProcessTree.

### 1. Config

This component reads the configuration file on initialization, and constructs the ProcessTree (empty of any pids, since everything remains unbooted for now).

### 2. Slave Monitor

This component is responsible for constantly attempting to boot any unbooted nodes in the ProcessTree. It is responsible for understanding
that child nodes must be started by issuing commands to their parents in the tree, and it is responsible for listening for messages about 
dead child processes, and restarting those processes (after killing their orphaned children).

### 3. Client Handler

This component is responsible for listening for connections from Clients (see docs/terminology if this seems confusing).
When a connection is received, this module issues a command to the Slave responsible for executing the requested Command,
and negotiates the socket pairing. It is also responsible to returning the command process exit status to the client.

### 4. File Monitor

The file monitor is responsible for listening to messages from slaves indicating features loaded, and inserting them in the process tree.
Additionally, it must interface with fsevents/inotify to watch for changes in these files, and kill processes (and all their children)
when a file that has been loaded is changed.
