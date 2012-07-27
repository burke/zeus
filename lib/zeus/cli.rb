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
    BANNER
  end
end
