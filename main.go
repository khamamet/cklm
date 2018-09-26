package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	Logger *log.Logger
	curDir string
)

func main() {

	var err error
	// get cur dir
	curDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// create log dir in the same dir
	logDir := filepath.Join(curDir, "logs")
	if err := os.MkdirAll(logDir, 0777); err != nil {
		panic(err.Error())
	}

	//Norml logging
	logfile, err := os.OpenFile(filepath.Join(curDir, "logs", "normal.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//if its possible to use third-party libs this logger is better
	/*
		log.SetOutput(&lumberjack.Logger{
			Filename:   filepath.Join(logDir, "normal.log"),
			MaxSize:    10, // megabytes
			MaxBackups: 7,
			MaxAge:     7, //days
		})
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	*/

	// create log for panic
	panicFileName := filepath.Join(logDir, "panic.log")
	var panicLog *os.File
	if _, err := os.Stat(panicFileName); err != nil {
		if panicLog, err = os.OpenFile(panicFileName, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644); err != nil {
			log.Println(e	rr.Error())
			return
		}
	} else {
		if panicLog, err = os.OpenFile(panicFileName, os.O_WRONLY|os.O_SYNC, 0644); err != nil {
			log.Println(err.Error())
			return
		}
	}
	syscall.Dup2(int(panicLog.Fd()), 1)
	syscall.Dup2(int(panicLog.Fd()), 2)

	//graceful shutdown
	gracefulStop := make(chan os.Signal, 2)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGALRM)
	go func() {
		for {
			sig := <-gracefulStop
			if sig == syscall.SIGHUP {
				log.Println("got signal SIGHUP. Ignored!")
			} else {
				log.Printf("got signal: %+v. Exiting...", sig)
				os.Exit(0)
			}
		}
	}()

	log.Println("process succsessfully started")
}
