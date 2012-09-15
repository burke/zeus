require "forwardable"

module Zeus
  module M
    ### Custom wrapper around an array of test methods
    # In charge of some smart querying, filtering, sorting, etc on the the
    # test methods
    class TestCollection
      include Enumerable
      extend Forwardable
      # This should act like an array, so forward some common methods over to the
      # internal collection
      def_delegators :@collection, :size, :<<, :each

      def initialize(collection = nil)
        @collection = collection || []
      end

      # Slice out tests that may be within the given line.
      # Returns a new TestCollection with the results.
      def within(line)
        # Into a new collection, filter only the tests that...
        self.class.new(select do |test|
          # are within the given boundary for this method
          # or include everything if the line given is nil (no line)
          line.nil? || (test.start_line..test.end_line).include?(line)
        end)
      end

      # Used to line up method names in `#sprintf` when `m` aborts
      def column_size
        # Boil down the collection of test methods to the name of the method's
        # size, then find the largest one
        @column_size ||= map { |test| test.name.to_s.size }.max
      end

      # Be considerate when printing out tests and pre-sort them by line number
      def by_line_number(&block)
        # On each member of the collection, sort by line number and yield
        # the block into the sorted collection
        sort_by(&:start_line).each(&block)
      end
    end
  end
end
