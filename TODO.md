## TODO (roughly prioritized)

* After an acceptor is killed, attempting to request that command while it is reloading causes a server error.
* Make sure that when a command process's connection is dropped, it is killed
* less leaky handling of at_exit pid killing
* Instead of exiting when requesting an as-yet-unbooted acceptor, wait until it's available then run.
* Refactor, refactor, refactor...
* Don't fork to handshake client to acceptor
* Eliminate the client-side exit lag for zeus commands.
* Support other frameworks?
* Figure out how to run full test suites without multiple env loads

## Ideas (not quite TODOs)

* (maybe) Start the preloader as a daemon transparently when any command is run, then wait for it to finish
* Support inotify on linux


