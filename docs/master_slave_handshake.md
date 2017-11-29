# Master/Worker Handshake

#### 1. Socket

The Worker is always started with an environment variable named `ZEUS_MASTER_FD`. The file descriptor at the given integer value is a socket to the Master process.

The Worker should open a UNIX Domain Socket using the `ZEUS_MASTER_FD` File Descriptor (`globalMasterSock`).

The Worker opens a new UNIX datagram Socketpair (`local`, `remote`)

The Worker sends `remote` across `globalMasterSock`.

#### 2. PID and Identifier

The Worker determines whether it has been given an Identifier. If it is the first-booted worker, it was booted
by the Master, and will not have one. When a Worker forks, it is passed an Identifier by the Master that it
passes along to the newly-forked process.

The Worker sends a "Pid & Identifier" message containing the pid and the identifier (blank if initial process)

#### 4. Action Result

The Worker now executes the code it's intended to run by looking up the action
in a collection of predefined actions indexed by identifier. In ruby this is implemented
as a module that responds to a method named according to each identifier.

If there were no runtime errors in evaluating the action, the Worker writes "OK" to `local`.

If there were runtime errors, the worker returns a string representing the errors in an arbitrary and
hopefully helpful format. It should normally be identical to the console output format should the errors
have been raised and printed to stderr.

Before the server kills a crashed worker process, it attempts to read
any loaded files from `local`, until that socket is closed.

#### 5. Loaded Files

Any time after the action has been executed, the Worker may (and should) send, over `local`, a list of files
that have been newly-loaded in the course of evaluating the action.

Languages are expected to implement this using clever tricks.

Steps 1-4 happend sequentially and in-order, but Submitting files in Step 5 should not prevent the Worker from
handling further commands from the master. The Worker should be considered 'connected' after Step 4.
