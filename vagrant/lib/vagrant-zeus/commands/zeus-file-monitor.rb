module VagrantPlugins::Zeus
  module Commands
    class FileMonitor < Vagrant.plugin(2, :command)
      # these are pointing to ext/ and buid/ in the gem root
      if /linux/ =~ RUBY_PLATFORM
        FILE_MONITOR_EXECUTABLE = File.join(
          File.dirname(__FILE__),
          '../../../ext/inotify-wrapper/inotify-wrapper'
        )
      else
        FILE_MONITOR_EXECUTABLE = File.join(
          File.dirname(__FILE__),
          '../../../build/fsevents-wrapper'
        )
      end

      class FileWatcher
        def initialize(machine)
          @machine = machine
          @file_monitor = spawn_file_monitor
          @zeus_connection = spawn_zeus_connection
        end

        def run
          modified_files_buf = ''
          watched_files_buf = ''
          ended = false

          while !ended
            ready = IO.select([@file_monitor, @zeus_connection])
            ready[0].each do |fh|
              if fh == @file_monitor
                begin
                  modified_files_buf += @file_monitor.readpartial(4096)
                rescue EOFError
                  puts "lost connection to the file monitor process, exiting!"
                  ended = true
                end
                modified_files_buf = process_modified_files(modified_files_buf)
              elsif fh == @zeus_connection
                begin
                  watched_files_buf += @zeus_connection.readpartial(4096)
                rescue EOFError
                  puts "lost connection to the zeus socket, exiting!"
                  ended = true
                end
                watched_files_buf = process_watched_files(watched_files_buf)
              end
            end
          end
        end

        private

        def spawn_file_monitor
          IO.popen(FILE_MONITOR_EXECUTABLE, 'r+')
        end

        def spawn_zeus_connection
          TCPSocket.new('localhost', @machine.config.zeus.file_monitor_port)
        end

        def process_modified_files(buf)
          lines = buf.sub(/.*\z/, '').split(/\n/)
          remaining = buf.sub(/\A.*\n/m, '')

          lines.each do |line|
            file = host_to_guest_path(line)
            if file
              @zeus_connection.write("#{file}\n")
            end
          end

          remaining
        end

        def process_watched_files(buf)
          lines = buf.sub(/.*\z/, '').split(/\n/)
          remaining = buf.sub(/\A.*\n/m, '')

          lines.each do |line|
            file = guest_to_host_path(line)
            if file
              @file_monitor.write("#{file}\n")
            end
          end

          remaining
        end

        def guest_to_host_path(path)
          guest_to_host_map.each do |guest, host|
            if path.start_with?(guest)
              return path.sub(/\A#{guest}/, host)
            end
          end

          # NOTE: we are explicitly not passing through things that don't live
          # on a synced folder, because they won't exist in a useful form on
          # the other side (and the inotify watcher has a slow poll for files
          # that don't exist)
          nil
        end

        def host_to_guest_path(path)
          host_to_guest_map.each do |host, guest|
            if path.start_with?(host)
              return path.sub(/\A#{host}/, guest)
            end
          end

          # NOTE: see above - this is less important because it doesn't trigger
          # the slow poll path, but still unnecessary
          nil
        end

        def guest_to_host_map
          paths = []
          @machine.config.vm.synced_folders.map do |_, options|
            if !options[:disabled]
              paths.push([
                options[:guestpath],
                File.absolute_path(options[:hostpath])
              ])
            end
          end
          paths.sort_by { |guest, host| guest.length }.reverse
        end

        def host_to_guest_map
          paths = []
          @machine.config.vm.synced_folders.map do |_, options|
            if !options[:disabled]
              paths.push([
                File.absolute_path(options[:hostpath]),
                options[:guestpath]
              ])
            end
          end
          paths.sort_by { |host, guest| host.length }.reverse
        end
      end

      def execute
        with_target_vms(nil, :single_target => true) do |machine|
          FileWatcher.new(machine).run
        end
        0
      end
    end
  end
end
