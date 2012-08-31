#import <Foundation/Foundation.h>
#include <CoreServices/CoreServices.h>
#include <sys/stat.h>
#include <fcntl.h>


static CFMutableArrayRef _watchedFiles;
static FSEventStreamRef  _activeStream;
static NSMutableDictionary *_fileIsWatched;

static int flagsWorthReloadingFor = \
kFSEventStreamEventFlagItemRemoved | \
kFSEventStreamEventFlagItemRenamed | \
kFSEventStreamEventFlagItemModified;

void myCallbackFunction(
                        ConstFSEventStreamRef streamRef,
                        void *clientCallBackInfo,
                        size_t numEvents,
                        void *eventPaths,
                        const FSEventStreamEventFlags eventFlags[],
                        const FSEventStreamEventId eventIds[])
{
    int i, flags;
    char **paths = eventPaths;
    
    for (i = 0; i < numEvents; i++) {
        flags = eventFlags[i];
        
        if (flags & (kFSEventStreamEventFlagItemIsFile | flagsWorthReloadingFor)) {
            printf("%s\n", paths[i]);
            fflush(stdout);
        }
    }
}

void configureStream()
{
    if (CFArrayGetCount(_watchedFiles) == 0) return;
    
    if (_activeStream) {
        FSEventStreamStop(_activeStream);
        FSEventStreamUnscheduleFromRunLoop(_activeStream, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
        //CFRelease(_activeStream);
    }
    
    _activeStream = FSEventStreamCreate(NULL,
                                        &myCallbackFunction,
                                        NULL,
                                        _watchedFiles,
                                        kFSEventStreamEventIdSinceNow,
                                        1.0, // latency
                                        kFSEventStreamCreateFlagFileEvents);
    
    FSEventStreamScheduleWithRunLoop(_activeStream, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
    
    FSEventStreamStart(_activeStream);
    
}

int maybeAddFileToWatchList(char *line)
{
    CFStringRef file = CFStringCreateWithCString(NULL, line, kCFStringEncodingASCII);
    struct stat buf;
    
    if ([_fileIsWatched valueForKey:(__bridge NSString *)file]) {
        return 0;
    } else if (stat(line, &buf) == 0) {
        [_fileIsWatched setValue:@"yes" forKey:(__bridge NSString *)file];
        CFArrayAppendValue(_watchedFiles, file);
        return 1;
    } else {
        return 0;
    }
}

void handleInputFiles()
{
    int anyChanges = 0;
    
    char line[2048];
    
    while (fgets(line, sizeof(line), stdin) != NULL) {
        line[strlen(line)-1] = 0;
        anyChanges |= maybeAddFileToWatchList(line);
    }
    
    if (anyChanges) {
        configureStream();
    }
}

void configureTimerAndRun()
{
    CFRunLoopTimerRef timer = CFRunLoopTimerCreate(NULL,
                                                   0,
                                                   0.5,
                                                   0,
                                                   0,
                                                   &handleInputFiles,
                                                   NULL);
    
    CFRunLoopAddTimer(CFRunLoopGetCurrent(), timer, kCFRunLoopDefaultMode);
    CFRunLoopRun();
}

int main(int argc, const char * argv[])
{
    int flags = fcntl(0, F_GETFL);
    flags |= O_NONBLOCK;
    fcntl(STDIN_FILENO, F_SETFL, flags);

    _watchedFiles = CFArrayCreateMutable(NULL, 0, NULL);
    _fileIsWatched = [[NSMutableDictionary alloc] initWithCapacity:500];
        
    configureTimerAndRun();
    return 0;
}