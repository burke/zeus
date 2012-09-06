# Message Format

There are a number of different types of messages passed between Master and Slave processes.

In the interest of simplifying Slave libraries, messages are sent as single packets over a UNIX datagram socket, with a single-letter prefix, followed by a colon, indicating the message type.

the parenthesesized values after each title are the message code, and the handling module.

#### Pid & Identifier message (`P`, `SlaveMonitor`)

This is sent from Slave to Master immediately after booting, to identify itself.

It is formed by joining the process's pid and identifier with a colon.

Example: `P:1235:default_bundle`

#### Action response message (`R`, `SlaveMonitor`)

This is sent from the Slave to the Master once the action has executed.

It can either be "OK", if the action was successful, or any other string, which should be a stderr-like 
representation of the error, including stack trace if applicable.

Example: `R:OK`

Example: `R:-e:1:in '<main>': unhandled exception`

#### Spawn Slave message (`S`, `SlaveMonitor`)

This is sent from the Master to the Slave and contains the Identifier of a new Slave to fork immediately.

Example: `S:test_environment`

#### Spawn Command message (`C`, `ClientHandler`)

This is sent from the Master to the Slave and contains the Identifier of a new Command to fork immediately.

Example: `C:console`

#### Client Command Request message (`Q`, `ClientHandler`)

This is sent from the (external) Client process to the ClientHandler. It contains the reqeusted command
identifier as well as any arguments (ie. the ARGV).

Example: `Q:testrb:-Itest -I. test/unit/module_test.rb`

#### Feature message (`F`, `FileMonitor`)

This is sent from the Slave to the Master to indicate it now depends on a file at a given path.

The path is expected to be the full, expanded path.

Example: `F:/usr/local/foo.rb`

