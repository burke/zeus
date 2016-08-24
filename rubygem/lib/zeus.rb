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

    def setup_master_socket!
      return master_socket if master_socket

      fd = ENV['ZEUS_MASTER_FD'].to_i
      self.master_socket = UNIXSocket.for_fd(fd)
    end

    def go(identifier=:boot)
      # Thanks to the magic of fork, this following line will return
      # many times: Every time the parent step receives a request to
      # run a command.
      if run_command = boot_steps(identifier)
        ident, local = run_command
        return command(ident, local)
      end
    end

    def boot_steps(identifier)
      # We need to use a mutex to serialize writes to the feature_pipe.
      # We don't need to syncronize any code before `run_action` though
      # because nothing else should be running.
      @feature_mutex = Mutex.new

      while true
        boot_step = catch(:boot_step) do
          $0 = "zeus slave: #{identifier}"

          setup_dummy_tty!
          master = setup_master_socket!
          feature_pipe_r, @feature_pipe = IO.pipe

          # I need to give the master a way to talk to me exclusively
          local, remote = UNIXSocket.pair(Socket::SOCK_STREAM)
          master.send_io(remote)

          # Now I need to tell the master about my PID and ID
          local.write "P:#{Process.pid}:#{@parent_pid || 0}:#{identifier}\0"
          local.send_io(feature_pipe_r)
          feature_pipe_r.close

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
              elsif code == "S"
                # Child, supposed to start another step:
                @parent_pid = forked_from

                @feature_pipe.close
                @feature_pipe = nil

                throw(:boot_step, ident.to_sym)
              else
                # Child, supposed to run a command:
                @parent_pid = forked_from

                @feature_pipe.close
                @feature_pipe = nil

                return [ident.to_sym, local]
              end
            end
          end
        end
        identifier = boot_step
      end
    end

    def notify_features(features)
      # Commands don't reload so they don't have a feature pipe
      return unless @feature_pipe
      @feature_mutex.synchronize do
        features.each do |t|
          @feature_pipe.puts(t)
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
      features, err = Zeus::LoadTracking.features_loaded_by do
        plan.after_fork unless identifier == :boot
        plan.send(identifier)
      end

      if err
        # If we received an error, report features to the master syncronously.
        # We need to do this before reporting the error to the master
        # otherwise it will kill us before we can report features.
        begin
          notify_features(features)
          @feature_pipe.close
        ensure
          report_error_to_master(socket, err)
        end
      else
        # If we booted successfully, report features in a new thread
        # so we can immediately begin listening for commands.
        socket.write "R:OK\0"
        Thread.new { notify_features(features) }
      end
    end
  end
end
