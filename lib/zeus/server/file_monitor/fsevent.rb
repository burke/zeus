require 'open3'
require 'pathname'

module Zeus
  class Server
    module FileMonitor
      class FSEvent
        WRAPPER_PATH = File.expand_path("../../../../../ext/fsevents-wrapper/fsevents-wrapper", __FILE__)

        def datasource          ; @io_out ; end
        def on_datasource_event ; handle_changed_files ; end
        def close_child_socket  ; end
        def close_parent_socket ; [@io_in, @io_out].each(&:close) ; end

        def initialize(&change_callback)
          @change_callback = change_callback
          @io_in, @io_out, _ = open_wrapper
          @givenpath_to_realpath = {}
          @realpath_to_givenpath = {}
          @buffer = ""
        end

        # The biggest complicating factor here is that ruby doesn't fully resolve
        # symlinks in paths, but FSEvents does. We resolve all paths fully with
        # Pathname#realpath, and keep mappings in both directions.
        # It's conceivable that the same file would be required by two different paths,
        # so we keep an array and fire callbacks for all given paths matching a real
        # path when a change is detected.
        def watch(given)
          return false if @givenpath_to_realpath[given]

          real = realpath(given)
          @givenpath_to_realpath[given] = real
          @realpath_to_givenpath[real] ||= []
          @realpath_to_givenpath[real] << given

          @io_in.write("#{real}\n")
          true
        end

        private

        def realpath(file)
          Pathname.new(file).realpath.to_s
        rescue Errno::ENOENT
          file
        end

        def open_wrapper
          Open3.popen2e(WRAPPER_PATH)
        end

        def handle_changed_files
          50.times { read_and_notify_files }
        rescue Stop
        end

        Stop = Class.new(Exception)

        def read_and_notify_files
          begin
            lines = @io_out.read_nonblock(1000)
          rescue Errno::EAGAIN
            raise Stop
          rescue EOFError
            Zeus.ui.error("fsevents-wrapper crashed.")
            Process.kill("INT", 0)
          end
          files = lines.split("\n")
          files[0] = "#{@buffer}#{files[0]}" unless @buffer == ""
          unless lines[-1] == "\n"
            @buffer = files.pop
          end

          files.each do |real|
            file_did_change(real)
          end
        end

        def file_did_change(real)
          realpaths_for_givenpath(real).each do |given|
            Zeus.ui.info("Dependency change at #{given}")
            @change_callback.call(given)
          end
        end

        def realpaths_for_givenpath(real)
          @realpath_to_givenpath[real] || []
        end

      end
    end
  end
end



