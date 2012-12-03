# Contributing to Zeus

Contributions to zeus are happily accepted, with one major caveat:

If you're adding support for some rarely-used library to the default `Zeus::Rails` plan, I'd prefer you release a separate gem called `zeus-foobar` to inject this functionality, rather than introducing extra complexity in the main project. Feel free to create a section in this project's README for extra features added this way.

# Hacking on Zeus' core

# Hacking on the Slave integration layer

## Getting up and running

Hacking on Zeus is perhaps not the most straightforward thing in the world. The development workflow could use a little bit of love.

### Installing Go

First prereq is

### Fetching dependencies

### Testing changes

Again, no real process for this. Here's what I do:

When I've changed the ruby code:

`make gem && gem install pkg/zeus-*.gem`

If I've only changed the Master/Client code (ie. anything written in Go):

`make darwin ; cp build/zeus-darwin-amd64 /path/to/gem_root/zeus-0.12.0/build`

I then test these changes manually by booting the application.

Unit test coverage is currently abysmal. I don't have a lot of experience testing stuff tied this closely to socket/terminal APIs, and haven't abstracted it far enough away that it's trivial. I plan to tackle this eventually. Whoever helps with this wins my eternal gratitude.
