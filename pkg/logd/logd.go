package logd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"time"
)

const (
	INFO    string = "INFO"
	DEBUG   string = "DEBUG"
	WARNING string = "* WARNING"
	ERROR   string = "** ERROR"
	FATAL   string = "*** FATAL ERROR"
	QUIT    string = "QUIT"
)

type Logd struct {
	loudLvls []string
	lg       *log.Logger
}

// make new logd object, holds io writers for log files and mongo client
func NewLogd(w io.Writer) *Logd {
	return &Logd{
		lg:       log.New(w, "", log.LstdFlags|log.Lshortfile),
		loudLvls: []string{INFO, WARNING, ERROR, FATAL, QUIT, DEBUG},
	}
}

// create and return the log file
// 01022006_150405
func GetLogWriter(existingF bool, path, timeFormat string) (io.Writer, error) {
	// print to stdout
	if path == "cli" {
		return os.Stdout, nil
	}
	// check if the file exists, if so return the file
	if !existingF {
		// create new file with formatted time (nothing if format "")
		var fName string = path
		if timeFormat != "" {
			fName += time.Now().Format("_" + timeFormat)
		}

		f, err := os.Create(fName)
		if err != nil {
			return nil, fmt.Errorf("failed to create file at %s\n**%w", fName, err)
		}
		return f, nil
	}

	// existing file, check if it exists, return open file if so
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("fatal: no file found at %s: %v", path, err)
		}
		return nil, fmt.Errorf("fatal: error occured for os.Stat(%s): %v", path, err)
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
}

// HIGH LEVEL FUNCS TO CALL IN SOURCE
// most general logging message, prints to main log file (lg)
func (l *Logd) Infof(msg string, args ...any) { l.log(INFO, msg, args...) }

// debugger level - only prints to dbg log file (qlg)
func (l *Logd) Debugf(msg string, args ...any) { l.log(DEBUG, msg, args...) }

// warnings, prints to main log file
func (l *Logd) Warnf(msg string, args ...any) { l.log(WARNING, msg, args...) }

// non fatal errors
func (l *Logd) Errorf(msg string, args ...any) { l.log(ERROR, msg, args...) }

// for graceful shutdowns - logs to main log file
func (l *Logd) Quitf(msg string, args ...any) { l.log(QUIT, msg, args...) }

// for fatal errors - mimics log.Fatalf behavior (log then exit)
func (l *Logd) Fatalf(msg string, args ...any) {
	l.log(FATAL, msg, args...)
	os.Exit(1)
}

// output to io writers associated with l -> all high level Infof etc funcs call this
func (l *Logd) log(level, msg string, args ...any) {
	prefix := fmt.Sprintf("[%s] ", level)
	l.lg.SetPrefix(prefix)

	var msgf string = msg
	if len(args) > 0 {
		msgf = fmt.Sprintf(msg, args...)
	}

	if slices.Contains(l.loudLvls, level) {
		if err := l.lg.Output(3, msgf); err != nil {
			l.lg.Printf("failed to output log msg %s", msgf)
		}
	}
}
