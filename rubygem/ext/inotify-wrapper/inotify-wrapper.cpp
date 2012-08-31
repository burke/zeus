#include <map>
#include <string>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/inotify.h>

#define EVENT_SIZE (sizeof (struct inotify_event))
#define EVENT_BUF_LEN (1024 * (EVENT_SIZE + 16))

using namespace std;

static int _inotify_fd;
static map<int, char*> _WatchedFiles;
static map<char*, bool> _FileIsWatched;

static int inotifyFlags = IN_ATTRIB | IN_MODIFY | IN_MOVE_SELF | IN_DELETE_SELF;

void maybeAddFileToWatchList(char *file)
{
  if (_FileIsWatched[file]) return;

  _FileIsWatched[file] = true;
  int wd = inotify_add_watch(_inotify_fd, file, inotifyFlags);
  _WatchedFiles[wd] = file;
}

void handleStdin()
{
  char line[2048];
  if (fgets(line, sizeof(line), stdin) == NULL) return;
  line[strlen(line)-1] = 0;

  maybeAddFileToWatchList(line);
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
    printf("%s\n", _WatchedFiles[event->wd]);
    fflush(stdout);

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
