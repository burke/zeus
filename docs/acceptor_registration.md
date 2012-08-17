# Acceptor Registration

When an acceptor is booted, it registers itself with the master process through UNIX Sockets. Specifically, it talks to the `AcceptorRegistrationMonitor`.

Here's an overview of the registration process:

1. `AcceptorRegistrationMonitor` creates a socketpair for Acceptor registration (`S_REG`)
2. When an `Acceptor` is started, it:
  1. Creates a new socketpair for communication with the master process (`S_ACC`)
  2. Sends one side of `S_ACC` over `S_REG` to the master.
  3. Sends its pid and then a newline character over `S_REG`.
3. `AcceptorRegistrationMonitor` receives first the IO and then the pid, and stores them for later reference.


