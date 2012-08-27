# Master/Slave Handshake

#### 1. Socket

The Slave is always started with an environment variable named `ZEUS_MASTER_FD`. The file descriptor at the given integer value is a socket to the Master process.

The Slave should open a UNIX Domain Socket using the `ZEUS_MASTER_FD` File Descriptor (`globalMasterSock`).

The Slave opens a new UNIX datagram Socketpair (`local`, `remote`)

The Slave sends `remote` across `globalMasterSock`.

#### 2. PID and Identifier

The Slave determines whether it has been given an Identifier. If it is the first-booted slave, it was booted
by the Master, and will not have one. When a Slave forks, it is passed an Identifier by the Master that it 
passes along to the newly-forked process.

The Slave sends a "Pid & Identifier" message containing the pid and the identifier (blank if initial process)

#### 4. Action Result

The Slave now executes the code it's intended to run by looking up the action
in a collection of predefined actions indexed by identifier. In ruby this is implemented
as a module that responds to a method named according to each identifier.

If there were no runtime errors in evaluating the action, the Slave writes "OK" to `local`.

If there were runtime errors, the slave returns a string representing the errors in an arbitrary and 
hopefully helpful format. It should normally be identical to the console output format should the errors
have been raised and printed to stderr.

#### 5. Loaded Files

Any time after the action has been executed, the Slave may (and should) send, over `local`, a list of files
that have been newly-loaded in the course of evaluating the action.

Languages are expected to implement this using clever tricks.

Steps 1-4 happend sequentially and in-order, but Submitting files in Step 5 should not prevent the Slave from
handling further commands from the master. The Slave should be considered 'connected' after Step 4.
