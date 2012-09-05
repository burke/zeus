# Slave Monitor Overview

The Slave Monitor is responsible for keeping each process in the process tree running.
It is responsible for the initial boot of all processes, and then subsequent maintenance,
by detecting when processes die, and restarting them.

### Messages Handled

* Pid & Identifier
* Action
* Action Response
* Spawn Slave
* Dead Child

## Components

The Slave Monitor has a few primary concerns:

* Booting a Slave
* Knowing when to boot a Slave

A slave should be booted when it is not running, but its parent is (or it is
the root state).

When a Slave has terminated, all of its children should be killed immediately,
with SIGKILL.

Booting a slave involves forking a process, negotiating sockets, pid, action,
and result. After this is done, the slave is ready.

The major problem is how to communicate this information about process states.

When a slave is booted, we can publish its identifer on a channel. A goroutine
will listen on this channel, look up all its child nodes, and start booting them.

When a slave dies, it will also be published on a channel. A goroutine will pull
the entry off the channel, kill all its children, and initiate rebooting.


