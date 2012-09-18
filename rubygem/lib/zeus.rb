# encoding: utf-8
require 'socket'
require 'json'

require 'zeus/load_tracking'

module Zeus
  class Plan
    def after_fork ; end
  end

  class << self
    attr_accessor :plan

    def go(identifier=:boot)
      $0 = "zeus slave: #{identifier}"
      # okay, so I ahve this FD that I can use to send data to the master.
      fd = ENV['ZEUS_MASTER_FD'].to_i
      master = UNIXSocket.for_fd(fd)

      # I need to give the master a way to talk to me exclusively
      local, remote = UNIXSocket.pair(Socket::SOCK_STREAM)
      master.send_io(remote)

      # Now I need to tell the master about my PID and ID
      local.write "P:#{Process.pid}:#{identifier}\0"

      # Now we run the action and report its success/fail status to the master.
      features = Zeus::LoadTracking.features_loaded_by {
        run_action(local, identifier)
      }

      # the master wants to know about the files that running the action caused us to load.
      Thread.new { notify_features(local, features) }

      # We are now 'connected'. From this point, we may receive requests to fork.
      loop do
        messages = local.recv(1024)
        messages.split("\0").each do |new_identifier|
          new_identifier =~ /^(.):(.*)/
          code, ident = $1, $2
          if code == "S"
            fork { plan.after_fork ; go(ident.to_sym) }
          else
            fork { plan.after_fork ; command(ident.to_sym, local) }
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

      begin
        pid_and_arguments = local.recv(1024)
        pid_and_arguments.chomp!("\0")
        # pid_and_arguments.force_encoding("ASCII-8BIT")
        File.open("b.log","a"){|f|f.puts "PA#{pid_and_arguments}" }
        pid_and_arguments =~ /(.*?):(.*)/
        client_pid, arguments = $1.to_i, $2
        arguments.chomp!("\0")
      rescue => e
        File.open("b.log","a"){|f|f.puts e.message ; f.puts e.backtrace}
      end

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
    end

    def kill_command_if_client_quits!(command_pid, client_pid)
      Thread.new {
        loop {
          begin
            File.open("b.log","a"){|f|f.puts "Checking #{client_pid}"}
            x=Process.kill(0, client_pid)
            File.open("b.log","a"){|f|f.puts "Checked #{client_pid} and got #{x}"}
          rescue Errno::ESRCH
            Process.kill(9, command_pid)
          end
          sleep 1
        }
      }
    end

    def notify_features(sock, features)
      features.each do |t|
        begin
          sock.write "F:#{t}\0"
        rescue Errno::ENOBUFS
          sleep 0.2
          retry
        end
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
      plan.send(identifier)
      socket.write "R:OK\0"
    rescue Exception => e
      report_error_to_master(socket, e)
    end

  end
end
