# encoding: utf-8
require 'socket'

# load exact json version from Gemfile.lock to avoid conflicts
gemfile = "#{ENV["BUNDLE_GEMFILE"] || "Gemfile"}.lock"
if File.exist?(gemfile) && version = File.read(gemfile)[/^    json \((.*)\)/, 1]
  gem 'json', version
end
require 'json'
require 'pty'
require 'set'

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
      master, self.dummy_tty = PTY.send(:open)
      Thread.new {
        loop { master.read(1024) }
      }
      STDIN.reopen(dummy_tty)
      STDOUT.reopen(dummy_tty)
    end

    def go(identifier=:boot)
      fd = ENV['ZEUS_MASTER_FD'].to_i
      sock = UNIXSocket.for_fd(fd)

      while true
        # This only returns after forking a child with a new
        # identifier and socket.
        identifier, sock = run_node(identifier, sock)
      end
    end

    def run_node(identifier, parent_sock)
      $0 = "zeus slave: #{identifier}"

      setup_dummy_tty!

      # I need to give the master a way to talk to me exclusively
      local, remote = UNIXSocket.pair(:DGRAM)
      parent_sock.send_io(remote)
      remote.close
      parent_sock.close

      # Now I need to tell the master about my PID and ID
      local.write "P:#{Process.pid}:#{@parent_pid || 0}:#{identifier}\0"

      feature_pipe_r, feature_pipe_w = IO.pipe
      local.send_io(feature_pipe_r)
      feature_pipe_r.close
      Zeus::LoadTracking.set_feature_pipe(feature_pipe_w)

      run_action(local, identifier)

      # We are now 'connected'. From this point, we may receive requests to fork.
      children = Set.new
      while true
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

          forked_from = Process.pid

          pid = fork
          if pid
            # We're in the parent. Record the child:
            children << pid
          else
            # Child
            @parent_pid = forked_from
            Zeus::LoadTracking.clear_feature_pipe

            if code == "S"
              # Boot another step
              return [ident.to_sym, local]
            else
              command(ident.to_sym, local)
              exit(0)
            end
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

      pid_and_argument_count = local.recv(2**16)
      pid_and_argument_count.chomp("\0") =~ /(.*?):(.*)/
      client_pid, argument_count = $1.to_i, $2.to_i
      arg_io = local.recv_io
      arguments = arg_io.read.chomp("\0").split("\0")

      if arguments.length != argument_count
        raise "Argument count mismatch: Expected #{argument_count}, got #{arguments.length}"
      end

      pid = fork {
        $0 = "zeus command: #{identifier}"

        plan.after_fork
        client_terminal = local.recv_io
        local.write "P:#{Process.pid}:#{@parent_pid}:\0"
        local.close

        $stdin.reopen(client_terminal)
        $stdout.reopen(client_terminal)
        $stderr.reopen(client_terminal)
        ARGV.replace(arguments)

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
      # Now we run the action and report its success/fail status to the master.
      begin
        Zeus::LoadTracking.track_features_loaded_by do
          plan.after_fork unless identifier == :boot
          plan.send(identifier)
        end

        socket.write "R:OK\0"
      rescue => err
        report_error_to_master(socket, err)
        raise
      end
    end
  end
end
