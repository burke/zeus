# Client/Coordinator/Command handshake

     Client    Coordinator    Command
    1  ---------->                | Command, Arguments, Pid
    2  ---------->                | Terminal IO
    3            ----------->     | Terminal IO
    4            ----------->     | Arguments, Pid
    5            <-----------     | pid
    6  <---------                 | pid
           (time passes)
    7            <-----------     | exit status
    8  <---------                 | exit status


#### 1. Command & Arguments (Client -> Coordinator)

The Coordinator always has a UNIX domain server listening at a known socket path.

The Client connects to this server and sends a string indicating the command to run and any arguments to run with (ie. the ARGV). See message_format.md for more info.

#### 2. Terminal IO (Client -> Coordinator)

The Client then sends an IO over the server socket to be used for raw terminal IO.

#### 3. Arguments (Coordinator -> Command)

The Coordinator sends the Client arguments from step 1 to the Command.

#### 4. Terminal IO (Coordinator -> Command)

The Coordinator forks a new Command process and sends it the Terminal IO from the Client.

#### 5. Pid (Command -> Coordinator)

The Command process sends the Coordinator its pid, using a Pid & Identifier message.

#### 6. Pid (Coordinator -> Client)

The Coordinator responds to the client with the pid of the newly-forked Command process.

The Client is now connected to the Command process.

#### 7. Exit status (Command -> Coordinator)

When the command terminates, it must send its exit code to the coordinator. This is normally easiest to implement as a wrapper process that does the setsid, then forks the command and `waitpid`s on it.

The form of this message is `{{code}}`, eg: `1`.

#### 8. Exit status (Coordinator -> Client)

Finally, the Coordinator forwards the exit status to the Client. The command cycle is now complete.

The form of this message is `{{code}}`, eg: `1`.

See [`message_format.md`](message_format.md) for more information on messages.

