open("Makefile", "wb") do |f|
  f.write <<-EOF
CXX = g++
CXXFLAGS = -O3 -g -Wall

file-listener: file-listener.o
	$(CXX) $(CXXFLAGS) $< -o $@

%.o: %.c
	$(CXX) $(CXXFLAGS) -c $< -o $@

install:
	# do nothing
  EOF
end
