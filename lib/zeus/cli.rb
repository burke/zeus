module Zeus::Cli
  extend self
  BANNER = <<-BANNER
usage: zeus <command> [<args>]

Commands:
  start
  testrb
  console
  server
  rake
  runner
  generate
BANNER

  def banner
    puts BANNER
  end

  def not_running(error)
    abort <<-ABORT
[#{error}]
Zeus may not running, try: 'zeus start'

#{BANNER}
    ABORT
  end

  def no_dot_zues(error)
    abort <<-ABORT
[#{error}]
Not .zues.rb found

#{BANNER}
ABORT
  end

  def start
    command = ARGV[0]

    if command == "start"
      dot_zeus
      server
    elsif command =~ /h|help/i or ARGV.empty?
      banner
    else
      client
    end
  end

  def dot_zeus
    require './.zeus.rb'
  rescue LoadError => e
    no_dot_zues(e)
    abort
  end

  def server
    Zeus::Server.run
  ensure
    Zeus::Client.cleanup!
  end

  def client
    Zeus::Client.run
  rescue Errno::ECONNREFUSED, Errno::ENOENT => e
    not_running(e)
    abort
  end
end
