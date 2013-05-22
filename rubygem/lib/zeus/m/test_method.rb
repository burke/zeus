module Zeus
  module M
    ### Simple data structure for what a test method contains.
    #
    # Too lazy to make a class for this when it's really just a bag of data
    # without any behavior.
    #
    # Includes the name of this method, what line on the file it begins on,
    # and where it ends.
    class TestMethod < Struct.new(:name, :start_line, :end_line)
      # Set up a new test method for this test suite class
      def self.create(suite_class, test_method, find_locations = true)
        # Hopefully it's been defined as an instance method, so we'll need to
        # look up the ruby Method instance for it
        method = suite_class.instance_method(test_method)

        if find_locations
          # Ruby can find the starting line for us, so pull that out of the array
          start_line = method.source_location.last

          # Ruby can't find the end line however, and I'm too lazy to write
          # a parser. Instead, `method_source` adds `Method#source` so we can
          # deduce this ourselves.
          #
          # The end line should be the number of line breaks in the method source,
          # added to the starting line and subtracted by one.
          end_line = method.source.split("\n").size + start_line - 1
        end

        # Shove the given attributes into a new databag
        new(test_method, start_line, end_line)
      end

      def escaped_name
        Regexp.escape(name)
      end
    end
  end
end
