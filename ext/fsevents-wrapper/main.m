//
//  main.m
//  fsevents-wrapper
//
//  Created by Burke Libbey on 2012-07-30.
//  Copyright (c) 2012 Burke Libbey. All rights reserved.
//

#import <Foundation/Foundation.h>
#include <CoreServices/CoreServices.h>
#include <sys/stat.h>


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
        }
        else { printf("%d\n", flags); }
    }
}

void configureStream()
{
    if (CFArrayGetCount(_watchedFiles) == 0) return;
    
    if (_activeStream) {
        FSEventStreamStop(_activeStream);
        FSEventStreamUnscheduleFromRunLoop(_activeStream, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
        CFRelease(_activeStream);
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
    
    fd_set fds;
    struct timeval tv;
    int retval;
    char line[2048];
    
    FD_ZERO(&fds);
    FD_SET(0, &fds);
    
    // Poll, don't block.
    tv.tv_usec = 0;
    tv.tv_sec = 0;
    
    retval = select(1, &fds, NULL, NULL, &tv);
    while (retval) {
        fgets(line, 2048, stdin);
        line[strlen(line)-1] = 0;
        
        anyChanges |= maybeAddFileToWatchList(line);
        
        retval = select(1, &fds, NULL, NULL, &tv);
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
    @autoreleasepool {
        _watchedFiles = CFArrayCreateMutable(NULL, 0, NULL);
        _fileIsWatched = [[NSMutableDictionary alloc] initWithCapacity:500];
        
        configureTimerAndRun();
    }
    return 0;
}

