package main

import (
	"os"
	"flag"
	"fmt"
	"strings"
	"time"
	"exec"

	l4g "log4go.googlecode.com/hg"
	mymy "github.com/ziutek/mymysql"
)

var (
	// general
	cmdname string        = os.Args[0]
	fs      *flag.FlagSet = flag.NewFlagSet(cmdname, flag.ExitOnError)
	//flags
	retryInterval *int    = fs.Int("r", 60, "retry interval, in second(s)")
	interval      *int    = fs.Int("t", 60, "sleep interval between two checks, in second(s)")
	skip          *bool   = fs.Bool("s", true, "whether skip error")
	mailAddrStr   *string = fs.String("m", "guchunjiang@pwrd.com,guchunjiang@wanmei.com", "mail addresses, delimited by ','")
	mailcmd       *string = fs.String("mail", "/usr/sbin/sendmail -t", "path to MTA")
	logFileName   *string = fs.String("f", os.Stderr.Name(), "output replication error log file")
	logLevelStr   *string = fs.String("l", "info", "log level filter(debug|info|warn|error)")
	batchMode     *bool   = fs.Bool("b", false, "execute once, ignore any intervals")
	// NID
	host string = "localhost"
	port string = "3306"
	user string
	pass string
	// logging
	log      = make(l4g.Logger)
	logLevel = l4g.INFO
	logFile  = os.Stderr
	// logic
	mailAddrs []string
	errorCount int
)

func parseFlags() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}()

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "    %v [OPTIONS] NID\n", cmdname)
		fmt.Fprintf(os.Stderr, "\n\nNID:\n")
		fmt.Fprintf(os.Stderr, "    \"h=?,P=?,u=?,p=?\"\n")
		fmt.Fprintf(os.Stderr, "\n\nOPTIONS:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[1:])
	// check log level
	switch *logLevelStr {
	case "debug":
		logLevel = l4g.INFO
	case "warn":
		logLevel = l4g.WARNING
	case "error":
		logLevel = l4g.ERROR
	default:
		logLevel = l4g.INFO
	}
	// check arg numbers
	if fs.NArg() != 1 {
		panic("no NID specified")
	}
	// check NID
	for _, pair := range strings.Split(fs.Arg(0), ",") {
		kv := strings.Split(pair, "=")
		switch kv[0] {
		case "h":
			host = kv[1]
		case "P":
			port = kv[1]
		case "u":
			user = kv[1]
		case "p":
			pass = kv[1]
		default:
			panic(fmt.Sprintf("mulformed NID: %v=%v", kv[0], kv[1]))
		}
	}
	// check log files
	if *logFileName != "/dev/stderr" {
		file, err := os.OpenFile(*logFileName,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Sprintf("%v\n", err))
		}
		logFile = file
	}
	// check mailing setting
	_, err := os.Lstat(strings.Split(*mailcmd, " ")[0])
	if err != nil {
		panic(fmt.Sprint(err))
	}
	mailAddrs = strings.Split(*mailAddrStr, ",")
}

func sendMail(content string) {
	tokens := strings.Split(*mailcmd, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	cmd.Start()
}

func isMySQLError(err *os.Error) bool {
	if _, ok := (*err).(mymy.Error); ok {
		return true
	}
	return false
}

func processRplStatus(mysql *mymy.MySQL) (slave bool, reconnect bool) {
	rows, res, err := mysql.Query("SHOW SLAVE STATUS")
	if err != nil {
		log.Warn("'SHOW SLAVE STATUS' returned with error: %v", err)
		return true, !isMySQLError(&err)
	}
	if len(rows) == 0 {
		log.Error("can't find slave info on this instance")
		return false, false
	}
	masterFile := strings.TrimSpace(
		rows[0].Str(res.Map["Relay_Master_Log_File"]))
	masterPos := strings.TrimSpace(rows[0].Str(res.Map["Exec_Master_Log_Pos"]))
	errorNo := strings.TrimSpace(rows[0].Str(res.Map["Last_Errno"]))
	lastError := strings.TrimSpace(rows[0].Str(res.Map["Last_Error"]))
	if lastError != "" {
		// log Last_Error
		msg := fmt.Sprintf("[%v %v] #%v: %v", 
			masterFile, masterPos, errorNo, lastError)
		log.Info(msg)
		sendMail(msg)
		// skip Last_Error
		_, _, err = mysql.Query(
			"SET GLOBAL SQL_SLAVE_SKIP_COUNTER = 1")
		if err != nil {
			log.Warn("trying to skip error but: %v", err)
			return true, !isMySQLError(&err)
		}
		_, _, err = mysql.Query("START SLAVE SQL_THREAD")
		if err != nil {
			log.Warn("trying to restart slave sql_thread but: %v", err)
			return true, !isMySQLError(&err)
		}
	}
	return true, false
}

func main() {
	parseFlags()
	log.AddFilter("stderr", logLevel,
		l4g.NewFormatLogWriter(logFile, "[%d %t] [%L] %M"))

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}()
	defer log.Close()

	// query slave status from MySQL instance
	mysql := mymy.New("tcp", "", host+":"+port, user, pass)
	defer mysql.Close()
	mysql.Debug = false
	if *batchMode {
		err := mysql.Connect()
		if err != nil {
			panic(fmt.Sprintf("can't connect to %v on port %v: %v",
				host, port, err))
		}
		processRplStatus(mysql)
	} else {
		log.Info("starting %v...", cmdname)
		for {
			if !mysql.IsConnected() {
				log.Info("connecting to %v port %v", host, port)
				err := mysql.Connect()
				if err != nil {
					log.Warn("can't connect to %v on port %v: %v",
						host, port, err)
					log.Warn("retry in %v seconds...", *retryInterval)
					// sleep for timeInterval minute(s)
					time.Sleep(int64(*retryInterval) * 1e9)
					continue
				}
				log.Info("connection established. start monitoring.")
			}
			slave, reconnect := processRplStatus(mysql)
			if !slave {
				// target server is not eligible to be monitored
				break
			}
			if reconnect {
				log.Warn("retry in %v seconds...", *retryInterval)
				mysql.Close()
				time.Sleep(int64(*retryInterval) * 1e9)
			} else {
				time.Sleep(int64(*interval) * 1e9)
			}
		}
		log.Info("closing...")
	}
}
