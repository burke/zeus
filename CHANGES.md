# Unreleased

https://github.com/burke/zeus/compare/v0.15.4...master

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
