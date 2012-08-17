module Zeus
  class ErrorPrinter
    attr_reader :error
    def initialize(error)
      @error = error
    end

    def write_to(io)
      io.puts "#{error.backtrace[0]}: #{error.message} (#{error.class})"
      error.backtrace[1..-1].each do |line|
        io.puts "\tfrom #{line}"
      end
    end

  end
end
