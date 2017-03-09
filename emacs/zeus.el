;;; zeus.el -- Minor mode for zeus application preloader

;;; Commentary:
;;
;; Implements `zeus-mode' for the zeus application preloader. This includes
;; support to start, stop and run commands against a running zeus server.

(require 's)
(require 'json)
(require 'cl)
(require 'dash)

(defcustom zeus-mode-completion-use-ido
  "Use ido mode for completion of zeus commands"
  t)

(defvar *zeus-server-process*
  "zeus server process buffer"
  nil)

(defun zeus--run (name buffer-name command &optional options)
  "Run a zeus command returning the running process buffer"
  (let ((zeus-command (s-join " " (list "zeus" (s-join " " options) command))))
    (process-buffer (start-process-shell-command name buffer-name zeus-command))))

(defun zeus--available-commands ()
  (let* ((zeus-json (json-read-from-string
                     (with-temp-buffer
                       (insert-file-contents "zeus.json")
                       (buffer-string))))
         (plan (alist-get 'plan zeus-json)))
    (mapcar 'symbol-name (zeus--extract-commands plan))))

(defun zeus--extract-commands (entries)
  (-flatten (delq nil (mapcar
                       (function (lambda (entry)
                                   (cond
                                    ((zeus--command-p entry) (car entry))
                                    ((listp entry) (zeus--extract-commands entry)))))
                       entries))))

(defun zeus--completing-read ()
  (if zeus-mode-completion-use-ido
      'ido-completing-read 'completing-read))

(defun zeus--command-p (entry)
  (and (listp entry) (arrayp (cdr entry))))

;;;###autoload
(defun zeus-start ()
  "Start the zeus server in the current directory using simple-status output"
  (interactive)
  (setq *zeus-server-process* (zeus--run "zeus-server" "*zeus-server*" "start" '("--simple-status")))
  (pop-to-buffer *zeus-server-process*))

(defun zeus-stop ()
  "Shutdown the currently active zeus server cleanly"
  (interactive)
  (when *zeus-server-process*
    (interrupt-process *zeus-server-process*)))

(defun zeus-run-command (cmd)
  "Run a command for zeus, i.e. zeus test to run the tests against the current zeus instance"
  (interactive
   (list (funcall (zeus--completing-read) "Command to run: " (zeus--available-commands))))
  (let ((cmd-name (concat "zeus" cmd)))
    (pop-to-buffer (zeus--run cmd-name "*zeus-command*" cmd))))

(provide 'zeus)
