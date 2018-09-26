package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var (
	ip       = "127.0.0.1"
	port     = 3333
	filename = "../data/data.csv"
)

type TData struct {
	Id            int
	Name          string
	Email         string
	Mobile_number string
}

func main() {

	var err error
	// get cur dir
	curDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	flag.StringVar(&filename, "f", "../golang/data/data.csv", "CSV filename to parse and load into DB")
	flag.IntVar(&port, "p", 3333, "TCP Port to connect to server")
	flag.StringVar(&ip, "i", "127.0.0.1", "ip address to connect to server")
	flag.Parse()
	// create log dir in the same dir
	logDir := filepath.Join(curDir, "logs")
	if err := os.MkdirAll(logDir, 0777); err != nil {
		panic(err.Error())
	}

	//Normal logging
	logfile, err := os.OpenFile(filepath.Join(curDir, "logs", "normal.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// create log for panic
	panicFileName := filepath.Join(logDir, "panic.log")
	var panicLog *os.File
	if _, err := os.Stat(panicFileName); err != nil {
		if panicLog, err = os.OpenFile(panicFileName, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644); err != nil {
			log.Println(err.Error())
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

	ReadAndLoad(filename)
}

func ReadAndLoad(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		//Parse data
		var dt TData
		spl := strings.Split(line, ",")
		if len(spl) == 4 {
			dt.Id, err = strconv.Atoi(spl[0])
			if err != nil {
				log.Println("Cannot parse id:", spl[0], "in the line:", line)
				continue
			}
			dt.Name = spl[1]
			dt.Email = spl[2]
			dt.Mobile_number = strings.Replace(strings.Replace(strings.Replace(spl[3], " ", "", -1), "(", "", -1), ")", "", -1)
			if strings.Index(dt.Mobile_number, "+44") != 0 {
				dt.Mobile_number = "+44" + dt.Mobile_number
			}
			//Call API
			err := dt.CallDBAPI()
			log.Println(err)
		} else {
			log.Println("Cannot parse the line:", line)
		}
	}

}
func (a *TData) CallDBAPI() error {
	c, err := net.Dial("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer c.Close()
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = c.Write(b)
	return err

}
