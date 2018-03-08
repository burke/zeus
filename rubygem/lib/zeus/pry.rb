# Only patch Pry if it is loaded by the app.
if defined?(Pry::Pager)
  class Pry::Pager
    def best_available
      # paging does not work in zeus so disable it
      NullPager.new(_pry_.output)
    end
  end
end
