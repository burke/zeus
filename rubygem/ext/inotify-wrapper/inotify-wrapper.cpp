#include <map>
#include <string>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/inotify.h>

#include <errno.h>

#define EVENT_SIZE (sizeof (struct inotify_event))
#define EVENT_BUF_LEN (1024 * (EVENT_SIZE + 16))

using namespace std;

static int _inotify_fd;
static map<int, string> _WatchedFiles;
static map<string, bool> _FileIsWatched;

// static int inotifyFlags = IN_ATTRIB | IN_MODIFY | IN_MOVE_SELF | IN_DELETE_SELF;
static int inotifyFlags = IN_ATTRIB | IN_MODIFY | IN_MOVE_SELF | IN_DELETE_SELF;

void maybeAddFileToWatchList(string file)
{
  if (_FileIsWatched[file]) return;

  int wd = inotify_add_watch(_inotify_fd, file.c_str(), inotifyFlags);
  int attempts = 0;
  // Files are momentarily inaccessible when they are rewritten. I couldn't
  // find a good way to deal with this, so we poll 'deleted' files for 0.25s or so
  // to see if they reappear.
  while (wd == -1 && errno == ENOENT) {
    usleep(10000);
    wd = inotify_add_watch(_inotify_fd, file.c_str(), inotifyFlags);
    if (attempts++ == 25) break; // try for at most about a quarter of a second
  }
  if (wd != -1) {
    _WatchedFiles[wd] = file;
    _FileIsWatched[file] = true;
  }
}

// This essentially removes a file from the watchlist then
// immediately re-adds it. This is because when a file is rewritten,
// as so many editors love to do, the watchdescriptor no longer refers to
// the file, so re must re-watch the path.
void replaceFileInWatchList(int wd, string file)
{
  _FileIsWatched.erase(file);
  _WatchedFiles.erase(wd);
  inotify_rm_watch(_inotify_fd, wd);
  maybeAddFileToWatchList(file);
}

void handleStdin()
{
  char line[2048];
  if (fgets(line, sizeof(line), stdin) == NULL) return;
  line[strlen(line)-1] = 0;

  maybeAddFileToWatchList(string(line));
}

void handleInotify()
{
  int length;
  int i = 0;
  char buffer[EVENT_BUF_LEN];
  string filename;

  length = read(_inotify_fd, buffer, EVENT_BUF_LEN);
  if (length < 0) return;

  while (i < length) {
    struct inotify_event *event = (struct inotify_event *) &buffer[i];
    string file = _WatchedFiles[event->wd];
    if (file != "") {
      printf("%s\n", file.c_str());
      fflush(stdout);
      replaceFileInWatchList(event->wd, file);
    }

    i += EVENT_SIZE + event->len;
  }
}

void go()
{
  fd_set rfds;
  int retval;

  for (;;) {
    FD_ZERO(&rfds);
    FD_SET(0, &rfds);
    FD_SET(_inotify_fd, &rfds);

    retval = select(_inotify_fd+1, &rfds, NULL, NULL, NULL);

    if (retval == -1) {
      // perror("select");
    } else if (retval) {
      if (FD_ISSET(0, &rfds))           handleStdin();
      if (FD_ISSET(_inotify_fd, &rfds)) handleInotify();
    }
  }
}


int main(int argc, const char *argv[])
{
  _inotify_fd = inotify_init();
  go();
}
