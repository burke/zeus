#include <set>
#include <string>
using namespace std;

#include <arpa/inet.h>
#include <errno.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <time.h>
#include <unistd.h>

set<string> watched_files;

void log(const char *fmt, ...)
{
    va_list ap;
    FILE *logfile;

    va_start(ap, fmt);

    vfprintf(stderr, fmt, ap);

    va_end(ap);
}

int start_server(int port)
{
    int sock, ssock, opt;
    struct sockaddr_in addr;
    socklen_t addrlen;

    sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock == -1) {
        log("error creating socket: %s\n", strerror(errno));
        return -1;
    }

    opt = 1;
    if (setsockopt(sock, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt))) {
        log("error setting socket options: %s\n", strerror(errno));
        close(sock);
        return -1;
    }

    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addrlen = sizeof(addr);
    if (bind(sock, (struct sockaddr *)&addr, addrlen)) {
        log("error binding to port %d: %s\n", port, strerror(errno));
        close(sock);
        return -1;
    }

    if (listen(sock, 0)) {
        log("error listening on socket: %s\n", strerror(errno));
        close(sock);
        return -1;
    }

    ssock = accept(sock, (struct sockaddr *)&addr, &addrlen);
    if (ssock == -1) {
        log("error while accepting a connection: %s\n", strerror(errno));
        close(sock);
        return -1;
    }
    close(sock);

    return ssock;
}

int handle_modified_file(int sock)
{
    ssize_t bytes;
    char buf[4096];

    bytes = recv(sock, buf, sizeof(buf) - 1, 0);
    if (bytes == -1) {
        log("error in recv: %s\n", strerror(errno));
    }
    else {
        buf[bytes] = '\0';
        if (bytes > 0 && buf[bytes - 1] == '\n') {
            buf[bytes - 1] = '\0';
        }
        if (watched_files.count(string(buf))) {
            log("*** reloading ***\n");
            printf("%s\n", buf);
            fflush(stdout);
        }
    }

    return bytes;
}

int handle_file_to_watch(int fd, int sock)
{
    char buf[4096];

    if (!fgets(buf, 4096, stdin)) {
        log("error in read: %s\n", strerror(errno));
        return 0;
    }
    else {
        size_t len;

        len = strlen(buf);
        if (len > 0) {
            if (send(sock, buf, len, 0) == -1) {
                log("failed to send '%s' to the socket: %s",
                    buf, strerror(errno));
            }
            buf[len - 1] = '\0';
            log("starting to watch '%s'\n", buf);
            watched_files.insert(string(buf));
        }

        return len;
    }
}

int main(int argc, char *argv[])
{
    int arg;
    unsigned short port;

    if (argc < 2) {
        log("usage: %s <port>\n", argv[0]);
        exit(1);
    }
    arg = atoi(argv[1]);
    if (arg <= 0 || arg >= 65536) {
        log("invalid port %s\n", argv[1]);
        log("usage: %s <port>\n", argv[0]);
        exit(1);
    }
    port = arg;

    for (;;) {
        int ssock;

        ssock = start_server(port);
        if (ssock == -1) {
            struct timespec tm;
            tm.tv_sec = 0;
            tm.tv_nsec = 100000000;
            nanosleep(&tm, NULL);
            continue;
        }

        for (;;) {
            fd_set fds;
            int ret;

            FD_ZERO(&fds);
            FD_SET(0, &fds);
            FD_SET(ssock, &fds);
            ret = select(ssock + 1, &fds, NULL, NULL, NULL);
            if (ret == -1) {
                continue;
            }
            else {
                if (feof(stdin)) {
                    log("stdin was closed, exiting");
                    exit(1);
                }
                if (FD_ISSET(0, &fds)) {
                    ret = handle_file_to_watch(0, ssock);
                    if (ret <= 0) {
                        log("lost stdin, exiting");
                        exit(1);
                    }
                }
                if (FD_ISSET(ssock, &fds)) {
                    ret = handle_modified_file(ssock);
                    if (ret <= 0) {
                        break;
                    }
                }
            }
        }

        close(ssock);
    }
}
