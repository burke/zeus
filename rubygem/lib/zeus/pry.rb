begin
  require "pry"

  class Pry::Pager
    def best_available
      # paging does not work in zeus so disable it
      NullPager.new(_pry_.output)
    end
  end

rescue LoadError => e
  # pry is not available, so no need to patch it
end
