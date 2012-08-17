# Client/Server handshake

This takes place in `lib/zeus/server/client_handler.rb`, `lib/zeus/client.rb`, and `lib/zeus/server/acceptor.rb`.

The model is kind of convoluted, so here's an explanation of what's happening with all these sockets:

## Running a command
1. ClientHandler has a UNIXServer (`SVR`) listening.
2. ClientHandler has a socketpair with the acceptor referenced by the command (see `docs/acceptor_registration.md`) (`S_ACC`)
3. When clienthandler receives a connection (`S_CLI`) on `SVR`:
  1. ClientHandler sends `S_CLI` over `S_ACC`, so the acceptor can communicate with the server's client.
  2. ClientHandler sends a JSON-encoded array of `arguments` over `S_ACC`
  3. Acceptor sends the newly-forked worker PID over `S_ACC` to clienthandler.
  4. ClientHandler forwards the pid to the client over `S_CLI`.


## A sort of network diagram
     client clienthandler acceptor
    1  ---------->                | {command: String, arguments: [String]}
    2  ---------->                | Terminal IO
    3            ----------->     | Terminal IO
    4            ----------->     | Arguments (json array)
    5            <-----------     | pid
    6  <---------                 | pid

