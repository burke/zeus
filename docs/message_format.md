# Message Format

There are a number of different types of messages passed between Master and Slave processes.

In the interest of simplifying Slave libraries, messages are send as single packets over a UNIX datagram socket,
with a single-letter prefix, followed by a colon, indicating the message type.

#### Pid & Identifier message (`P`)

This is sent from Slave to Master immediately after booting, to identify itself.

It is formed by joining the process's pid and identifier with a colon.

Example: `P:1235:default_bundle`

#### Action message (`A`)

This is sent from the Master to the Slave, and contains the action code to execute.

Example: `A:require 'rails/all'\nBundler.require(:default)\n`

#### Action response message (`R`)

This is sent from the Slave to the Master once the action has executed.

It can either be "OK", if the action was successful, or any other string, which should be a stderr-like 
representation of the error, including stack trace if applicable.

Example: `R:OK`

Example: `R:-e:1:in '<main>': unhandled exception`

#### Spawn Slave message (`S`)

This is sent from the Master to the Slave and contains the Identifier of a new Slave to fork immediately.

Example: `S:test_environment`

#### Spawn Command message (`C`)

This is sent from the Master to the Slave and contains the Identifier of a new Command to fork immediately.

Example: `C:console`

#### Dead child message (`D`)

This is sent from the Slave to the Master when one of its child processes has terminated.

Example: `D:1234`

#### Feature message (`F`)

This is send frome the Slave to the Master to indicate it now depends on a file at a given path.

The path is expected to be the full, expanded path.

Example: `F:/usr/local/foo.rb`

