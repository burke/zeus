if /linux/ =~ RUBY_PLATFORM
  open("Makefile", "wb") do |f|
    f.write <<-EOF
CXX = g++
CXXFLAGS = -O3 -g -Wall

inotify-wrapper: inotify-wrapper.o
	$(CXX) $(CXXFLAGS) $< -o $@

%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

install:
	# do nothing
    EOF
  end
else
  open("Makefile", "wb") do |f|
    f.write <<-EOF
install:
	# do nothing
    EOF
  end
end
