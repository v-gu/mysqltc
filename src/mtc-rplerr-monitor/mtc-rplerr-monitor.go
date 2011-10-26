package main

import (
	"os"
	"os/signal"
	"io/ioutil"
	"flag"
	"fmt"
	"strings"
	"time"
	"exec"

	l4g "log4go.googlecode.com/hg"
	mymy "github.com/ziutek/mymysql"
)

const (
	IO_ERROR  = "IO_ERROR"
	SQL_ERROR = "SQL_ERROR"
)

var (
	// general
	cmdname string        = os.Args[0]
	fs      *flag.FlagSet = flag.NewFlagSet(cmdname, flag.ExitOnError)
	//flags
	retryInterval *int    = fs.Int("r", 60, "retry interval, in second(s)")
	interval      *int    = fs.Int("t", 60, "sleep interval between two checks, in second(s)")
	skip          *bool   = fs.Bool("s", true, "whether skip error")
	mailAddrStr   *string = fs.String("m", "vincent.gu@perfectworld.com", "mail addresses, delimited by ','")
	mailcmd       *string = fs.String("mail", "/usr/bin/sendmail -t", "path to MTA")
	mailSendGap   *int    = fs.Int("g", 30, "how many retries before send out remider mail with the same topic")
	logFileName   *string = fs.String("e", os.Stderr.Name(), "general log filename")
	sqlogFilename *string = fs.String("f", os.Stdout.Name(), "sql error log filename")
	logLevelStr   *string = fs.String("l", "info", "log level filter(debug|info|warn|error)")
	batchMode     *bool   = fs.Bool("b", false, "execute once, ignore any intervals")
	pidfileName   *string = fs.String("pidfile", "", "file existed only when program was running, with PID filled in")
	// NID
	host string = "localhost"
	port string = "3306"
	user string
	pass string = "pass"
	// logging
	log        = make(l4g.Logger)
	logLevel   = l4g.INFO
	logFile    = os.Stderr
	sqlog      = make(l4g.Logger)
	sqlogLevel = l4g.INFO
	sqlogFile  = os.Stdout
	// logic
	hostname      string
	errorCount    int
	masterHost    string
	masterPort    string
	gsid          uint64                  // global sequence id
	errorStatuses map[string]*ErrorStatus = make(map[string]*ErrorStatus, 2)
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
		logLevel = l4g.DEBUG
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
}

type RplError struct {
	errType string
	errno   string
	error   string
	logFile string
	pos     string
}

func (err *RplError) string() string {
	return fmt.Sprintf("[%v %v] #%v: %v",
		err.logFile, err.pos, err.errno, err.error)
}

type ErrorStatus struct {
	sid         uint64 // sequence id
	rplError    *RplError
	repeatCount int
	msg         string // problem resolve message
}

func sendmail(content string) {
	tokens := strings.Split(*mailcmd, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Warn("allocate stdin for sendmail failed: %v", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Warn("failed to allocate stdout for sendmail: %v", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Warn("failed to allocate stderr for sendmail: %v", err)
		return
	}
	// format header
	var header, signature string
	header += fmt.Sprintf("To: %v\n", *mailAddrStr)
	header += fmt.Sprintf("Subject: MySQL replication error on [%v:%v]\n",
		host, port)
	header += fmt.Sprintf("From: /%v/mtc-rplerr-monitor\n", hostname)
	header += fmt.Sprintf("Date: %v\n", time.LocalTime())
	header += fmt.Sprintf("\n")
	header += fmt.Sprintf("Error detected on MySQL replication chain "+
		"%v:%v -> %v:%v\n", masterHost, masterPort, host, port)
	signature += fmt.Sprintf("\n-- \nRegards,\nmtc-rplerr-monitor\n")
	signature += fmt.Sprintf("PLESASE DO NOT REPLY DIRECTLY TO THIS EMAIL")
	log.Debug("mail:\n%v%v%v", header, content, signature)
	stdin.Write([]byte(header))
	stdin.Write([]byte(content))
	stdin.Write([]byte(signature))
	stdin.Close()
	cmd.Start()
	// acquire outputs
	out, err := ioutil.ReadAll(stdout)
	stdout.Close()
	if err != nil {
		log.Warn("failed to read STDOUT from sendmail: %v", err)
	}
	error, err := ioutil.ReadAll(stderr)
	stderr.Close()
	if err != nil {
		log.Warn("failed to read STDERR from sendmail: %v", err)
	}
	if out != nil && len(out) > 0 {
		log.Info(string(out))
	}
	if error != nil && len(error) > 0 {
		log.Warn(string(error))
	}
}

func isMySQLError(err *os.Error) bool {
	if _, ok := (*err).(mymy.Error); ok {
		return true
	}
	return false
}

func processRplStatus(mysql *mymy.MySQL) (slave bool, reconnect bool) {
	slave, reconnect = true, false
	rows, res, err := mysql.Query("SHOW SLAVE STATUS")
	if err != nil {
		log.Warn("'SHOW SLAVE STATUS' returned with error: %v", err)
		reconnect = !isMySQLError(&err)
		return
	}
	if len(rows) == 0 {
		log.Error("can't find slave info on this instance")
		slave, reconnect = false, false
		return
	}
	gsid += 1
	log.Debug("current gsid: %v", gsid)
	masterHost = strings.TrimSpace(rows[0].Str(res.Map["Master_Host"]))
	if masterHost == "127.0.0.1" {
		masterHost = host
	}
	masterPort = strings.TrimSpace(rows[0].Str(res.Map["Master_Port"]))
	var rplErrors []*RplError
	// check IO error
	errNo := strings.TrimSpace(rows[0].Str(res.Map["Last_IO_Errno"]))
	if errNo != "0" {
		error := strings.TrimSpace(rows[0].Str(res.Map["Last_IO_Error"]))
		file := strings.TrimSpace(rows[0].Str(res.Map["Master_Log_File"]))
		pos := strings.TrimSpace(rows[0].Str(res.Map["Read_Master_Log_Pos"]))
		rplError := &RplError{IO_ERROR, errNo, error, file, pos}
		rplErrors = append(rplErrors, rplError)
		log.Debug("add IO error: %v", rplError)
	}
	// check SQL error
	errNo = strings.TrimSpace(rows[0].Str(res.Map["Last_SQL_Errno"]))
	if errNo != "0" {
		error := strings.TrimSpace(rows[0].Str(res.Map["Last_SQL_Error"]))
		file := strings.TrimSpace(rows[0].Str(res.Map["Relay_Master_Log_File"]))
		pos := strings.TrimSpace(rows[0].Str(res.Map["Exec_Master_Log_Pos"]))
		rplError := &RplError{SQL_ERROR, errNo, error, file, pos}
		rplErrors = append(rplErrors, rplError)
		log.Debug("add SQL error: %v", rplError)
	}
	// check repetition
	for _, rplError := range rplErrors {
		if prevErr := errorStatuses[rplError.errType]; prevErr == nil {
			errorStatus := &ErrorStatus{gsid, rplError, 0, ""}
			errorStatuses[rplError.errType] = errorStatus
		} else {
			if (gsid-prevErr.sid) > 1 || // fell too far behind
				rplError.pos != prevErr.rplError.pos ||
				rplError.logFile != prevErr.rplError.logFile {
				prevErr.rplError = rplError
				prevErr.repeatCount = 0
			} else {
				prevErr.repeatCount += 1
			}
			prevErr.sid = gsid // set sid up-to-date
			prevErr.msg = ""
		}
	}
	// deal with the situation
	for errorType, errorStatus := range errorStatuses {
		rplError := errorStatus.rplError
		if errorStatus.sid != gsid {
			log.Debug("do not process %v [%v %v] because its obsoleted",
				errorType, rplError.logFile, rplError.pos)
			continue
		}
		log.Debug("Processing %v [%v %v]",
			errorType, rplError.logFile, rplError.pos)
		if errorStatus.repeatCount == 0 {
			log.Info("found rpl error: #%v@[%v %v]",
				rplError.errno, rplError.logFile, rplError.pos)
		}
		if *skip {
			if errorType == SQL_ERROR {
				log.Info("trying to skip rpl error: #%v@[%v %v]",
					rplError.errno, rplError.logFile, rplError.pos)
				_, _, err = mysql.Query(
					"SET GLOBAL SQL_SLAVE_SKIP_COUNTER = 1")
				if err != nil {
					msg := fmt.Sprintf("trying to skip error but: %v, will "+
						"retry later.", err)
					log.Warn(msg)
					errorStatus.msg = msg
					reconnect = reconnect || !isMySQLError(&err)
					continue
				}
				_, _, err = mysql.Query("START SLAVE SQL_THREAD")
				if err != nil {
					msg := fmt.Sprintf("trying to restart slave sql_thread "+
						"but: %v, will retry later", err)
					log.Warn(msg)
					errorStatus.msg = msg
					reconnect = reconnect || !isMySQLError(&err)
					continue
				}
			}
		}
		if errorType == IO_ERROR {
			msg := fmt.Sprintf("IO_ERROR cannot only be resolved manually or " +
				"by itself.\n")
			log.Warn(msg)
			errorStatus.msg = msg
			continue
		}
	}
	// format mail contents
	var mail string
	for errorType, errorStatus := range errorStatuses {
		rplError := errorStatus.rplError
		if errorStatus.sid != gsid {
			log.Debug("do not process %v [%v %v] because its obsoleted",
				errorType, rplError.logFile, rplError.pos)
			continue
		}
		if errorStatus.repeatCount%*mailSendGap != 0 {
			continue
		}
		log.Debug("formatting mail for %v [%v %v]",
			errorType, rplError.logFile, rplError.pos)
		mail += fmt.Sprintf("\n%v:\n", errorType)
		mail += fmt.Sprintf("  - WARNING: %v\n", rplError.string())
		if errorStatus.msg != "" {
			mail += fmt.Sprintf("  - %v\n", errorStatus.msg)
		} else {
			if *skip {
				mail += fmt.Sprintf("  - Note: this error was jumped and " +
					"logged.\n")
			} else {
				mail += fmt.Sprintf("  - WARNING: this error was logged " +
					"but still blocking the replication, manual override " +
					"is required.\n")
			}
		}
	}
	if mail != "" {
		go sendmail(mail)
	}
	return
}

func createPidfile() {
	if *pidfileName != "" {
		log.Debug("creating pidfile: %v", *pidfileName)
		pidfile, err := os.OpenFile(*pidfileName,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(fmt.Sprintf("Cannot create PID file:%v", err))
		}
		fmt.Fprint(pidfile, os.Getpid())
		pidfile.Close()
		log.Debug("pidfile %v created", *pidfileName)
	}
}
func removePidfile() {
	err := os.Remove(*pidfileName)
	if err != nil {
		log.Debug("failed to remove pidfile %v: %v", *pidfileName, err)
	} else {
		log.Debug("pidfile %v created", *pidfileName)
	}
}

func exit(err interface{}) {
	if err != nil {
		log.Error(err)
		log.Info("aborting...")
	} else {
		log.Info("stopping...")
	}
	removePidfile()
	sqlog.Close()
	log.Close()
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func main() {
	// parse command line flags
	parseFlags()
	// init logger
	log.AddFilter("stderr", logLevel,
		l4g.NewFormatLogWriter(logFile, "[%d %t] [%L] %M"))
	defer log.Close()
	sqlog.AddFilter("stdout", sqlogLevel,
		l4g.NewFormatLogWriter(sqlogFile, "[%d %t] %M"))
	defer sqlog.Close()
	// global error catching
	defer exit(recover())
	createPidfile()
	// spawn os signal handler
	go func() {
		for {
			switch sig := (<-signal.Incoming).(os.UnixSignal); sig {
			case os.SIGINT, os.SIGHUP, os.SIGQUIT, os.SIGTERM, os.SIGKILL:
				{
					exit(fmt.Sprintf("%v received", sig))
				}
			}
		}
	}()
	// main
	hostname, _ = os.Hostname()
	// query slave status from MySQL instance
	mysql := mymy.New("tcp", "", host+":"+port, user, pass)
	defer mysql.Close()
	mysql.Debug = false
	log.Info("starting %v...", cmdname)
	if *batchMode {
		err := mysql.Connect()
		if err != nil {
			panic(fmt.Sprintf("can't connect to %v on port %v: %v",
				host, port, err))
		}
		processRplStatus(mysql)
	} else {
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
	}
	exit(nil)
}
