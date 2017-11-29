# Terminology

* a Client is a process initiated by the user requesting zeus to run a command.

* the Master is the Go program which mediates all the interaction between the other processes

* a Worker is a process managed by Zeus which is used to load dependencies for commands

* a Command process is one forked from a Worker and connected to a Client
