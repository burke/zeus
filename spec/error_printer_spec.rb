require 'spec_helper'

require 'stringio'

module Zeus
  describe ErrorPrinter do

    let(:error) {
      begin
        raise
      rescue => e
        e
      end
    }

    it 'prints an error just like ruby does by default' do
      io = StringIO.new
      ErrorPrinter.new(error).write_to(io)
      io.rewind
      lines = io.readlines
      lines[0].should =~ /^[^\s]*error_printer_spec.rb:\d+:in `.+':  \(RuntimeError\)$/
      lines[1].should =~ /^\tfrom .*\.rb:\d+:in `.*'$/
      lines.size.should > 5
    end

  end
end
