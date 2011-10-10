package main

import (
	"os"
	"flag"
	"fmt"
	"strings"
	"strconv"
	"time"

	l4g "log4go.googlecode.com/hg"
	mymy "github.com/ziutek/mymysql"
)

var (
	cmdname string      = os.Args[0]
	fs      *flag.FlagSet = flag.NewFlagSet(cmdname, flag.ExitOnError)
	//flags
	host          *string = fs.String("h", "localhost", "MySQL host name")
	port          *int    = fs.Int("P", 3306, "MySQL port")
	user          *string = fs.String("u", "root", "login username")
	pass          *string = fs.String("p", "", "login password")
	retryInterval *int    = fs.Int("r", 60, "retry interval, in second(s)")
	interval      *int    = fs.Int("t", 60, "sleep interval between two checks, in second(s)")
	skip          *bool   = fs.Bool("s", true, "whether skip error")
	mailaddr      *string = fs.String("m", "guchunjiang@pwrd.com,guchunjiang@wanmei.com", "mail addresses, delimited by ','")
	mailcmd       *string = fs.String("mail", "/usr/sbin/sendmail -t", "path to MTA")
	logFile       *string = fs.String("f", os.Stderr.Name(), "output replication error log file")
	batchMode     *bool   = fs.Bool("b", true, "execute once, ignore any intervals")

	// logging
	log = make(l4g.Logger)

	errorCount int
)

func parseFlags() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Arg parsing failed: %v\n", err)
		}
	}()

	fs.Usage = func() {
		fmt.Printf("Usage:\n")
		fmt.Printf("    %v [OPTIONS] NID\n", cmdname)
		fmt.Printf("\n\nNID:\n")
		fmt.Printf("    h=?,P=?,u=?,p=?\n")
		fmt.Printf("\n\nOPTIONS:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[1:])
	if fs.NArg() == 0 && fs.NFlag() == 0 {
		
	}
}

func checkRplStatus(mysql *mymy.MySQL) (shutdown bool) {
	rows, res, err := mysql.Query("SHOW SLAVE STATUS")
	if err != nil {
		log.Error("'SHOW SLAVE STATUS' returned with error: %v\n", err)
		return false
	}
	if len(rows) == 0 {
		log.Warn("this MySQL instance is not a slave!")
		log.Warn("exiting...")
		return true
	}
	lastError := strings.TrimSpace(rows[0].Str(res.Map["Last_Error"]))
	if lastError != "" {
		// log Last_Error
		msg := fmt.Sprintf("[%s] %s\n", time.LocalTime().String(),
			lastError)
		if *logFile == os.Stdout.Name() {
			os.Stdout.WriteString(msg)
		} else {
			file, err := os.OpenFile(*logFile,
				os.O_WRONLY|os.O_APPEND|os.O_CREATE,
				0666)
			if err != nil {
				log.Error("error opening log file: %v\n", err)
				return true
			}
			err = file.Close()
			if err != nil {
				log.Error("error closing log file: %v\n", err)
				return true
			}
		}
		// skip Last_Error
		_, _, err = mysql.Query(
			"SET GLOBAL SQL_SLAVE_SKIP_COUNTER = 1")
		if err != nil {
			log.Error("trying to skip error but: %v\n", err)
			return false
		}
		_, _, err = mysql.Query("START SLAVE SQL_THREAD")
		if err != nil {
			log.Error("trying to restart slave sql_thread but: %v\n", err)
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
	parseFlags()

	// query slave status from MySQL instance
	mysql := mymy.New("tcp", "", *host+":"+strconv.Itoa(*port), *user, *pass)
	mysql.Debug = false
	for {
		if !mysql.IsConnected() {
			log.Info("connecting to %v port %v\n", host, port)
			err := mysql.Connect()
			if err != nil {
				log.Error("can't connect to %v on port %v: %v\n", host, port, err)
				log.Warn("retry in %v seconds...", retryInterval)
				// sleep for timeInterval minute(s)
				time.Sleep(int64(*retryInterval) * 1e9)
				continue
			} else {
				log.Info("connection established. start monitoring.\n")
			}
		}
		if checkRplStatus(mysql) {
			break
		}
		time.Sleep(int64(*interval) * 1e9)
	}
	log.Info("shutting down...")
	mysql.Close()
}
