require 'socket'
require 'json'
require './fake_zeus'

def report_error_to_master(local, e)
  local.write "R:FAIL"
end

def run_action(socket, identifier)
  FakeZeus.send(identifier)
  socket.write "R:OK"
rescue Exception => e
  File.open("wtf.log", "a") {|f| f.puts e.inspect , e.backtrace}
  report_error_to_master(socket, e)
end

def notify_newly_loaded_files
end

def handle_dead_children
  loop do
    pid = Process.wait(-1, Process::WNOHANG)
    break if pid.nil?
    local.send("D:#{pid}")
  end
rescue Errno::ECHLD
end

def go(identifier=:boot)
  identifier = identifier.to_sym
  $0 = "zeus slave: #{identifier}"
  # okay, so I ahve this FD that I can use to send data to the master.
  fd = ENV['ZEUS_MASTER_FD'].to_i
  master = UNIXSocket.for_fd(fd)

  # I need to give the master a way to talk to me exclusively
  local, remote = UNIXSocket.pair(:DGRAM)
  master.send_io(remote)

  # Now I need to tell the master about my PID and ID
  local.write "P:#{Process.pid}:#{identifier}"

  # Now we run the action and report its success/fail status to the master.
  run_action(local, identifier)

  # the master wants to know about the files that running the action caused us to load.
  Thread.new { notify_newly_loaded_files }

  trap("CHLD") { handle_dead_children(local) }

  # We are now 'connected'. From this point, we may receive requests to fork.
  loop do
    new_identifier = local.recv(1024)
    File.open("wtf.log", "a") {|f| f.puts "GOT #{new_identifier}"}
    if new_identifier =~ /^S:/
      fork { go(new_identifier.sub(/^S:/,'')) }
    else
      fork { command(new_identifier.sub(/^C:/,''), local) }
    end
  end

end

def command(identifier, sock)
  local, remote = UNIXSocket.pair(:DGRAM)
  sock.send_io(remote)

  arguments = local.recv(1024)
  File.open("wtf.log", "a") {|f| f.puts arguments.inspect }
  client_terminal = local.recv_io

  local.write "P:#{Process.pid}:\n"

  $stdin.reopen(client_terminal)
  $stdout.reopen(client_terminal)
  $stderr.reopen(client_terminal)
  ARGV.replace(JSON.parse(arguments))

  # Process.setsid

  FakeZeus.send(identifier)

  puts "HOMG"

end

__FILE__ == $0 and go()
