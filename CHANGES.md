### 0.13.1

* Handle the `shared_connection` hack that some people do to share an AR connection between threads (77d2b9bfc67977dfd8c17eed03fe2a8a25870c11)

* Improved a few cases where client processes disconnect unexpectedly.

* Changed up the slave/master IPC, solving a bunch of issues on Linux, by switching from a socket to a pipe.

* Client terminations are now handled a bit more gracefully. The terminal is always reset from raw mode, and the cursor is reset to column 0.
