require 'fileutils'

version = File.read('../../VERSION').chomp

File.open('zeusversion.go', 'w') { |f| f.puts <<END
package zeusversion

const VERSION string = "#{version}"
END
}
