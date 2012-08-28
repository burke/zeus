module FakeZeus
  class << self
    def method_missing(s, *a)
      File.open("omg.log", "a") { |f| f.puts "FakeZeus.#{s}" }
      if s.to_sym == :console
        exec "htop"
      elsif s.to_sym == :development_environment
        # raise "HOMG"
      end
    end
  end
end

