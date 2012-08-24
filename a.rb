require 'socket'

fd = ENV['ZEUS_MASTER_FD'].to_i
a,b = UNIXSocket.pair
sock = UNIXSocket.for_fd(fd)
sock.send_io(a)
b << "Hello from ruby!"
b.close
