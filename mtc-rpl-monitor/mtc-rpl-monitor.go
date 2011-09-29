package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"strings"
	"strconv"
	"time"
	mymy "github.com/ziutek/mymysql"
)

var (
	host          string
	port          uint
	user          string
	pass          string
	retryInterval int
	interval      int
	skip          bool
	mailaddr      string
	mailcmd       string
	file          string // output log file

	errorCount int
)

func init() {
	// parse command-line flags
	flag.StringVar(&host, "h", "localhost", "MySQL host name")
	flag.UintVar(&port, "P", 3306, "MySQL port")
	flag.StringVar(&user, "u", "root", "login username")
	flag.StringVar(&pass, "p", "", "login password")
	flag.IntVar(&retryInterval, "r", 60, "retry interval, in second(s)")
	flag.IntVar(&interval, "t", 60,
		"sleep interval between two checks, in second(s)")
	flag.BoolVar(&skip, "s", true, "whether skip error")
	flag.StringVar(&mailaddr, "m",
			// "Don.Nguyen@perfectworld.com,guchunjiang@pwrd.com",
			"guchunjiang@pwrd.com,guchunjiang@wanmei.com",
			"mail addresses, delimited by ','")
	flag.StringVar(&mailcmd, "mail", "/usr/sbin/sendmail -t",
		"path to MTA")
	flag.StringVar(&file, "f", os.Stdout.Name(),
		"output replication error log file")
	flag.Parse()
	if flag.NArg() == 0 && flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func checkRplStatus(mysql *mymy.MySQL) (shutdown bool) {
	rows, res, err := mysql.Query("SHOW SLAVE STATUS")
	if err != nil {
		log.Printf("[ERROR] 'SHOW SLAVE STATUS' returned with error: %v\n", err)
		return false ""
	}
	if len(rows) == 0 {
		log.Println("[WARN] this MySQL instance is not a slave!")
		log.Println("[WARN] exiting...")
		return true 
	}
	lastError := strings.TrimSpace(rows[0].Str(res.Map["Last_Error"]))
	if lastError != "" {
		// log Last_Error
		msg := fmt.Sprintf("[%s] %s\n", time.LocalTime().String(),
			lastError)
		if file == os.Stdout.Name() {
			os.Stdout.WriteString(msg)
		} else {
			logFile, err := os.OpenFile(file,
				os.O_WRONLY|os.O_APPEND|os.O_CREATE,
				0666)
			if err != nil {
				log.Printf("[ERROR] error opening log file: %v\n", err)
				return true
			}
			err = logFile.Close()
			if err != nil {
				log.Printf("[ERROR] error closing log file: %v\n", err)
				return true
			}
		}
		// skip Last_Error
		_, _, err = mysql.Query(
			"SET GLOBAL SQL_SLAVE_SKIP_COUNTER = 1")
		if err != nil {
			log.Printf("trying to skip error but: %v\n", err)
			return false
		}
		_, _, err = mysql.Query("START SLAVE SQL_THREAD")
		if err != nil {
			log.Printf("trying to restart slave sql_thread but: %v\n", err)
			return false
		}
	}
	return false
}

func complain() {
}

func sendMail(addr string, content string) {
	
}

func main() {
	// query slave status from MySQL instance
	mysql := mymy.New("tcp", "", host+":"+strconv.Uitoa(port), user, pass)
	mysql.Debug = false
	for {
		if !mysql.IsConnected() {
			log.Printf("[INFO] connecting to %v port %v\n",
				host, port)
			err := mysql.Connect()
			if err != nil {
				log.Printf("[ERROR] can't connect to %v on port "+
					"%v: %v\n", host, port, err)
				log.Printf("retry in %v seconds...", retryInterval)
				// sleep for timeInterval minute(s)
				time.Sleep(int64(retryInterval) * 1e9)
				continue
			} else {
				log.Printf("[INFO] connection established. start monitoring.\n")
			}
		}
		if checkRplStatus(mysql) {
			break
		}
		time.Sleep(int64(interval) * 1e9)
	}
	log.Printf("[INFO] shutting down...")
	mysql.Close()
}
