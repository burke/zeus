require 'zeus'

describe Zeus do
  class MyTestPlan < Zeus::Plan
    def self.mutex
      @mutex ||= Mutex.new
    end

    def self.boot
      require_relative 'assets/boot'
      Thread.new do
        mutex.synchronize do
          Zeus::LoadTracking.add_feature(File.join(__dir__, 'assets', 'boot_delayed.rb'))
        end
      end
    end
  end

  Zeus.plan = MyTestPlan

  before do
    MyTestPlan.mutex.lock
  end

  after do
    MyTestPlan.mutex.unlock if MyTestPlan.mutex.locked?
  end

  context 'booting' do
    before do
      # Don't reopen STDOUT
      Zeus.dummy_tty = true
    end

    it 'boots and tracks features' do
      master_r, master_w = UNIXSocket.pair(Socket::SOCK_STREAM)
      ENV['ZEUS_MASTER_FD'] = master_w.to_i.to_s

      thr = Thread.new do
        begin
          Zeus.go
        rescue Interrupt
          next
        rescue => e
          STDERR.puts "Zeus terminated with exception: #{e.message}"
          STDERR.puts e.backtrace.map {|line| " #{line}"}
        end
      end

      begin
        # Receive the control IO and start message
        ctrl_io = master_r.recv_io(UNIXSocket)
        begin
          # We use recv instead of readline on the UNIXSocket to avoid
          # converting it to a buffered reader. That seems to interact
          # badly with passing file descriptors around on Linux.
          proc_msg = "P:#{Process.pid}:0:boot\0"
          expect(ctrl_io.recv(proc_msg.length)).to eq(proc_msg)

          feature_io = ctrl_io.recv_io

          ready_msg = "R:OK\0"
          expect(ctrl_io.recv(ready_msg.length)).to eq(ready_msg)
          begin
            # We should receive the synchronously required feature immediately
            expect(feature_io.readline).to eq(File.join(__dir__, 'assets', 'boot.rb') + "\n")

            # We should receive the delayed feature after unlocking its mutex
            MyTestPlan.mutex.unlock
            expect(feature_io.readline).to eq(File.join(__dir__, 'assets', 'boot_delayed.rb') + "\n")
          ensure
            feature_io.close
          end
        ensure
          ctrl_io.close
        end
      ensure
        thr.raise(Interrupt.new)
      end
    end

    it 'tracks features after booting has completed' do
    end
  end
end
