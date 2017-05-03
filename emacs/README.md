# Zeus.el

Package to run zeus commands and manage the zeus server from inside Emacs

## Installation

Put `zeus.el` inside your load path (`~/.emacs.d/site-lisp` for example)

```elisp
(setq site-lisp-dir (expand-file-name "site-lisp" user-emacs-directory))
(add-to-list 'load-path site-lisp-dir)
(require 'zeus)
```
## Usage

### `zeus-start` command

Starts a new zeus process in the current directory, zeus needs to support
`--simple-status` for this to work.

### `zeus-stop` command

Stop the process created via `zeus-start`

### `zeus-run-command` command

run an available zeus command in the current directory, with additional
arguments as requested. With `C-u` prefix this will be run as the compile
command, otherwise it will be run in a new buffer. 

Commands are completed using `ido` completion if available.

