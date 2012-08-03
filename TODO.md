## TODO (roughly prioritized)

* Make sure that when a command process's connection is dropped, it is killed
* less leaky handling of at_exit pid killing
* Refactor, refactor, refactor...
* Support other frameworks?
* Figure out how to run full test suites without multiple env loads

## Ideas (not quite TODOs)

* (maybe) Start the preloader as a daemon transparently when any command is run, then wait for it to finish
* Support inotify on linux


