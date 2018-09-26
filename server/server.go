package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
)

var (
	port   = 3333
	ip     = "127.0.0.1"
	config TConfig
	lSrcDB *sql.DB
)

type TData struct {
	Id            int
	Name          string
	Email         string
	Mobile_number string
}

func main() {
	flag.IntVar(&port, "p", 3333, "TCP Port to start the server")
	flag.StringVar(&ip, "i", "127.0.0.1", "ip address to start the server")
	flag.Parse()

	var err error
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

	config, err = LoadConfig(filepath.Join(curDir, "config.ini"))

	//Connect To MySQL DB
	connstring := config.Server.SrcUserName + ":" + config.Server.SrcPassword + "@" + config.Server.SrcNetwork +
		"(" + config.Server.SrcAddress + ")/" + config.Server.DBName + "?charset=" + config.Server.Charset
	//log.Println(connstring)
	lSrcDB, err = sql.Open("mysql", connstring)
	if err != nil {
		log.Fatalln("ERROR mySQL server doesnt exist! ", err)
	}

	defer lSrcDB.Close()
	if err = lSrcDB.Ping(); err != nil {
		log.Fatalln("Cannot connect to DB! Check parameters in ini file")
	}

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

	//CREATE TABLE IN THE DB
	CreateTables(lSrcDB)

	laddr, err := net.ResolveTCPAddr("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}
func handleRequest(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for {
		ok := scanner.Scan()
		b := scanner.Bytes()
		if !ok {
			log.Println("Reached EOF on server connection.")
			break
		}
		var data TData
		if err := json.Unmarshal(b, &data); err != nil {
			log.Println("Reached EOF on server connection.")
			break
		} else {
			SaveData(lSrcDB, data)
		}
	}
	conn.Close()
}
func CreateTables(lSrcDB *sql.DB) error {
	asql := "CREATE TABLE IF NOT EXISTS `ciklum`.`rawdata` ( " +
		" `id` INT NOT NULL, " +
		" `name` VARCHAR(45) NOT NULL, " +
		" `email` VARCHAR(128) NOT NULL, " +
		" `mobile_number` VARCHAR(12) NOT NULL, " +
		" PRIMARY KEY (`id`), " +
		" UNIQUE INDEX `id_UNIQUE` (`id` ASC));"
	_, err := lSrcDB.Exec(asql)
	return err
}
func SaveData(lSrcDB *sql.DB, data TData) error {
	//Should use Prepare
	asql := "INSERT INTO `ciklum`.`rawdata` (id, name, email, mobile_number) VALUES (" +
		strconv.Itoa(data.Id) + "," +
		"'" + escapeSQL(data.Name) + "'," +
		"'" + escapeSQL(data.Email) + "'," +
		"'" + escapeSQL(data.Mobile_number) + "')" +
		" ON DUPLICATE KEY UPDATE " +
		"name = '" + escapeSQL(data.Name) + "'," +
		"email = '" + escapeSQL(data.Email) + "'," +
		"mobile_number = '" + escapeSQL(data.Mobile_number) + "'"
	_, err := lSrcDB.Exec(asql)
	return err
}
func escapeSQL(src string) string {
	src = strings.Replace(src, "\\", "\\\\", -1)
	src = strings.Replace(src, "'", "\\'", -1)
	src = strings.Replace(src, "\"", "\\\"", -1)
	return src
}
