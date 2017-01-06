# Unreleased

https://github.com/burke/zeus/compare/v0.15.11...master

# 0.15.11

https://github.com/burke/zeus/compare/v0.15.10...v0.15.11

* Send TERM instead of KILL when killing processes to allow them time
  to clean up after themselves.
* Add support for reporting files loaded after an action is completed.
* Add file change information to trace logs for debugging and move unexpected
  process logging to trace logs.
* Moved zeus client error message to stderr so they are more visible

# 0.15.10

https://github.com/burke/zeus/compare/v0.15.9...v0.15.10

* Revert changes that required Zeus to be in the Gemfile (Reverts #530 to fix #570).
* Substantially rework the Zeus state machine to avoid race conditions that cause hangs.

# 0.15.9

https://github.com/burke/zeus/compare/v0.15.8...v0.15.9

* Fix critical bug in status output that prevents booting in some environments.
  https://github.com/burke/zeus/issues/567
* Debounce status chart updates to avoid duplicate lines

# 0.15.8

https://github.com/burke/zeus/compare/v0.15.6...v0.15.8

* Replace file change monitoring with native Go code. This means the
  zeus Gem no longer requires native extensions and file monitoring is
  much faster and more reliable.
* Track files from exceptions during Zeus actions in Ruby.
* Fix a thread safety in SlaveNode state access.

# 0.15.7

*0.15.7 was never really released due to a critical bug in OS X file monitoring.*

# 0.15.6

https://github.com/burke/zeus/compare/v0.15.5...v0.15.6

* Better output and error recovery for Vagrant plugin gem
* Fixed zeus gem monkey patch that made `Kernel#load` public

# 0.15.5

https://github.com/burke/zeus/compare/v0.15.4...v0.15.5

* Inject Rails console helpers when using Pry (@leods92)
* Integrate better with autorun capabilities of test frameworks (Minitest, Rspec) (@latortuga, @blowmage, @zenspider)
* Help resolve issues loading improper gem versions by simplifying load tracking and leaning on bundler (@kgrz)

# 0.15.4

https://github.com/burke/zeus/compare/v0.15.3...v0.15.4

* Fix issues invoking `zeus test` without arguments when using RSpec (@sshao)
* Prevent infinite loop in inotify plugin (@e2)

# 0.15.3

https://github.com/burke/zeus/compare/v0.15.2...v0.15.3

* Add support for RSpec 3.0+ out of the box (hat tip @PaBLoX-CL, @maxcal, @stabenfeldt, #308, #474)

# 0.15.2

https://github.com/burke/zeus/compare/v0.15.1...v0.15.2

* Add support for Minitest 5.0+ (@rcook)
* Add `--config` to command line interface for specifying your own zeus.json file (@nevir)
* Don't double escape Regex generated for test names (@grosser)
* Add back Linux make command (make linux) for building Linux-only binaries

# 0.15.1

https://github.com/burke/zeus/compare/v0.15.0...v0.15.1

* Revert IRB is not a module change in favor of supporting standard rails
  console and Pry usage. `zeus console` now works out of the box as long as you
  aren't resorting to reassinging the IRB constant to Pry. If you are doing
  something like that, please create your own custom plan and zeus.json file to
  workaround the errors you receive.

# 0.15.0

https://github.com/burke/zeus/compare/v0.14.0.rc1...v0.15.0

* Reworked the build process. Makefile completely rewritten, cross-compilation
  now done with [gox](github.com/mitchellh/gox).

* Removed cucumber from default plan.

* Bugfix: Read JSON gem version from Gemfile.lock to avoid version conflicts

* Feature: Add TEST_HELPER environment variable to force Rspec or (#393, @despo)

* Bugfix PTY.open issue with Rubinius

* Bugfix: IRB is not a module (@bendilley)

### 0.14.0.rc1

*0.14.0 was never really released due to bugs and other reasons.*

https://github.com/burke/zeus/compare/v0.13.3...v0.14.0.rc1

### *Undocumented Period*

### 0.13.3

https://github.com/burke/zeus/compare/v0.13.2...v0.13.3

* Returns correct status code [#252](https://github.com/burke/zeus/issues/252)

* Other bug fixes

### 0.13.2

https://github.com/burke/zeus/compare/v0.13.1...v0.13.2

### 0.13.1

* Handle the `shared_connection` hack that some people do to share an AR
  connection between threads (77d2b9bfc67977dfd8c17eed03fe2a8a25870c11)

* Improved a few cases where client processes disconnect unexpectedly.

* Changed up the slave/master IPC, solving a bunch of issues on Linux, by
  switching from a socket to a pipe.

* Client terminations are now handled a bit more gracefully. The terminal is
  always reset from raw mode, and the cursor is reset to column 0.
