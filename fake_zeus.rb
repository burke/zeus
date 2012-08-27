module FakeZeus
  class << self
    def method_missing(s, *a)
      File.open("omg.log", "a") { |f| f.puts "FakeZeus.#{s}" }
    end
  end
end

