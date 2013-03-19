# encoding: utf-8
require 'socket'
require 'json'
require 'pty'

require 'zeus/load_tracking'
require 'zeus/plan'
require 'zeus/version'

module Zeus
  class << self
    attr_accessor :plan, :dummy_tty, :master_socket

    # this is totally asinine, but readline gets super confused when it's
    # required at a time when stdin or stdout is not connected to a TTY,
    # no matter what we do to tell it otherwise later. So we create a dummy
    # TTY in case readline is required.
    #
    # Yup.
    def setup_dummy_tty!
      return if self.dummy_tty
      master, self.dummy_tty = PTY.open
      Thread.new {
        loop { master.read(1024) }
      }
      STDIN.reopen(dummy_tty)
      STDOUT.reopen(dummy_tty)
    end

    def setup_master_socket!
      return master_socket if master_socket

      fd = ENV['ZEUS_MASTER_FD'].to_i
      self.master_socket = UNIXSocket.for_fd(fd)
    end

    def go(identifier=:boot)
      $0 = "zeus slave: #{identifier}"

      setup_dummy_tty!
      master = setup_master_socket!
      feature_pipe_r, feature_pipe_w = IO.pipe

      # I need to give the master a way to talk to me exclusively
      local, remote = UNIXSocket.pair(Socket::SOCK_STREAM)
      master.send_io(remote)

      # Now I need to tell the master about my PID and ID
      local.write "P:#{Process.pid}:#{identifier}\0"
      local.send_io(feature_pipe_r)

      # Now we run the action and report its success/fail status to the master.
      features = Zeus::LoadTracking.features_loaded_by {
        run_action(local, identifier)
      }

      # the master wants to know about the files that running the action caused us to load.
      Thread.new { notify_features(feature_pipe_w, features) }

      # We are now 'connected'. From this point, we may receive requests to fork.
      children = Set.new
      loop do
        messages = local.recv(2**16)

        # Reap any child runners or slaves that might have exited in
        # the meantime. Note that reaping them like this can leave <=1
        # zombie process per slave around while the slave waits for a
        # new command.
        children.each do |pid|
          children.delete(pid) if Process.waitpid(pid, Process::WNOHANG)
        end

        messages.split("\0").each do |new_identifier|
          new_identifier =~ /^(.):(.*)/
          code, ident = $1, $2
          pid = nil
          if code == "S"
            children << fork { go(ident.to_sym) }
          else
            children << fork { command(ident.to_sym, local) }
          end
        end
      end
    end

    private

    def command(identifier, sock)
      $0 = "zeus runner: #{identifier}"
      Process.setsid

      local, remote = UNIXSocket.pair(:DGRAM)
      sock.send_io(remote)
      remote.close
      sock.close

      pid_and_arguments = local.recv(2**16) # if starting client before boot slave is latched, we get stuck here. We also get stuck here in #182.
      pid_and_arguments.chomp!("\0")
      pid_and_arguments =~ /(.*?):(.*)/
      client_pid, arguments = $1.to_i, $2
      arguments.chomp!("\0")

      pid = fork {
        $0 = "zeus command: #{identifier}"

        plan.after_fork
        client_terminal = local.recv_io
        local.write "P:#{Process.pid}:\0"
        local.close

        $stdin.reopen(client_terminal)
        $stdout.reopen(client_terminal)
        $stderr.reopen(client_terminal)
        ARGV.replace(JSON.parse(arguments))

        plan.send(identifier)
      }

      kill_command_if_client_quits!(pid, client_pid)

      Process.wait(pid)
      code = $?.exitstatus || 0

      local.write "#{code}\0"

      local.close
    rescue Exception
      # If anything at all went wrong, kill the client - if anything
      # went wrong before the runner can clean up, it might hang
      # around forever.
      Process.kill(:TERM, client_pid)
    end

    def kill_command_if_client_quits!(command_pid, client_pid)
      Thread.new {
        loop {
          begin
            Process.kill(0, client_pid)
          rescue Errno::ESRCH
            Process.kill(9, command_pid)
            exit 0
          end
          sleep 1
        }
      }
    end

    def notify_features(pipe, features)
      features.each do |t|
        pipe.puts t
      end
    end

    def report_error_to_master(local, error)
      str = "R:"
      str << "#{error.backtrace[0]}: #{error.message} (#{error.class})\n"
      error.backtrace[1..-1].each do |line|
        str << "\tfrom #{line}\n"
      end
      str << "\0"
      local.write str
    end

    def run_action(socket, identifier)
      plan.after_fork unless identifier == :boot
      plan.send(identifier)
      socket.write "R:OK\0"
    rescue Exception => e
      report_error_to_master(socket, e)
    end

  end
end
