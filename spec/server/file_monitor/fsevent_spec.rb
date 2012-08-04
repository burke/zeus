require 'socket'
require 'tempfile'
require 'fileutils'
require 'securerandom'

require 'zeus'

module Zeus::Server::FileMonitor
  describe FSEvent do

    let(:fsevent) { FSEvent.new() { } }

    it 'registers files to be watched' do
      _, io_out = stub_open_wrapper!

      fsevent.watch("/a/b/c.rb")
      io_out.readline.chomp.should == "/a/b/c.rb"
    end

    it 'only registers a file with the wrapper script once' do
      _, io_out = stub_open_wrapper!

      files = ["a", "a", "b", "a", "b", "c", "d", "a"]
      files.each { |f| fsevent.watch(f) }

      files.uniq.each do |file|
        io_out.readline.chomp.should == file
      end
    end

    it 'passes changed files to a callback' do
      io_in, io_out = stub_open_wrapper!

      # to prove that very long filenames aren't truncated anywhere:
      filename = SecureRandom.hex(8000) + ".rb"

      results = []
      fsevent = FSEvent.new { |f| results << f }

      io_in.puts filename
      fsevent.stub(realpaths_for_givenpath: [filename])
      # test that the right socket is used, and it's ready for reading.
      IO.select([fsevent.datasource])[0].should == [io_out]

      Zeus.ui.should_receive(:info).with(%r{#{filename}})
      fsevent.on_datasource_event
      results[0].should == filename
    end


    it 'closes sockets not necessary in child processes' do
      io_in, io_out = stub_open_wrapper!
      fsevent.close_parent_socket

      io_in.should be_closed
      io_out.should be_closed
    end

    it 'integrates with the wrapper script to detect changes' do
      results = []
      callback = ->(path){ results << path }
      fsevent = FSEvent.new(&callback)

      file = Tempfile.new('fsevent-test')

      fsevent.watch(file.path)

      Zeus.ui.should_receive(:info).with(%r{#{file.path}})

      FileUtils.touch(file.path)
      IO.select([fsevent.datasource], [], [], 3)[0] # just wait for the data to appear
      fsevent.on_datasource_event
      results[0].should == file.path

      file.unlink
    end

    private

    def stub_open_wrapper!
      io_in, io_out = Socket.pair(:UNIX, :STREAM)
      FSEvent.any_instance.stub(open_wrapper: [io_in, io_out])

      [io_in, io_out]
    end

  end
end
