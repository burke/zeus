module FakeZeus
  class << self
    def method_missing(s, *a)
      sleep 1.5
      File.open("omg.log", "a") { |f| f.puts "FakeZeus.#{s}" }
    end
  end
end

