// mtc-syncdump will make a still image of a leaf MySQL instance with
// point-in-time informations regarding to all preceding instances.
// 
// For example, using mtc-syndump to make an still image of a random MySQL
// instance N3, which itself is inside a replication chain:
//
//     N1 <- N2 <- N3 <- N4   (N3 replicate from N2, etc...)
//
// Then with the login informations of N1 and N3 and some proper privilege
// setups on them, mtc-syncdump can obtain the dump with point-in-time binlog
// positions of every preceding instances, like:
// 
//     N1: mysql-bin.000005 48471238 "2011-09-12 12:32:49"
//     N2: mysql-bin.000023 12389772 "2011-09-12 19:21:43"
//     N3: mysql-bin.000001 21383457 "2011-09-12 08:11:53"
//
// In which case, one can restore this backup with log sequences and time in
// mind.
package main

import (
	"os"
	"flag"
	"fmt"
	l4g "log4go.googlecode.com/hg"
	mymysql "github.com/ziutek/mymysql"
)

var (
	// args:
	cmdname string        = os.Args[0]
	fs      *flag.FlagSet = flag.NewFlagSet(cmdname, flag.ExitOnError)
	// args: general
	verbose   *bool = fs.Bool("v", false, "verbose output")
	debugMode *bool = fs.Bool("debug", false, "debug mode")
	// args: set info for upsteam topmost MySQL instance
	rootHost *string = fs.String("rh", "", "host of upstream topmost MySQL instance")
	rootPort *int    = fs.Int("rP", 3306, "port of upstream topmost MySQL instance")
	rootUser *string = fs.String("ru", "rman", "username of upstream topmost MySQL instance")
	rootPass *string = fs.String("rp", "", "password of upstream topmost MySQL instance")
	// args: set info for leaf MySQL instance
	leafHost *string = fs.String("h", "localhost", "host of target MySQL instance")
	leafPort *int    = fs.Int("P", 3306, "port of target MySQL instance")
	leafUser *string = fs.String("u", "rman", "username of target MySQL instance")
	leafPass *string = fs.String("p", "", "password of target MySQL instance")
	// args: dumping elements
	dumpDb    *string = fs.String("d", "", "\"db1,db2,...\", include only these db")
	dumpDbExc *string = fs.String("D", "", "\"db1,db2,...\", exclude these db")
	dumpTb    *string = fs.String("t", "", "\"db1.tb1,db1.tb2,...\", include only these tables")
	dumpTbExc *string = fs.String("T", "", "\"db1.tb1,db1.tb2,...\", exclude these tables")

	// logging controls
	log = make(l4g.Logger)
)

func ParseArgs() {
	fs.Usage = func() {
		fmt.Printf("Usage of %v:\n", cmdname)
		fmt.Printf("  %v [ROOT_INSTANCE] BACKUP_INSTANCE [OPTIONS] > backup_file.sql\n\n",
			cmdname)
		fmt.Printf("\nROOT_INSTANCE:\n")
		fmt.Printf("  -rh=HOST [-rP=PORT] [-ru=USER] [-rp=PASS]\n\n")
		fmt.Printf("\nBACKUP_INSTANCE:\n")
		fmt.Printf("  -h=HOST [-P=PORT] [-u=USER] [-p=PASS]\n\n")
		fmt.Printf("\nOPTION:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[1:])
	if fs.NArg() > 0 || (fs.NArg() == 0 && fs.NFlag() == 0) {
		log.Error("wrong args")
		fs.Usage()
		os.Exit(1)
	}
}

func main() {
	log.AddFilter("stderr", l4g.DEBUG,
		l4g.NewFormatLogWriter(os.Stderr, "[%d %t] [%L] %M"))
	defer log.Close()
	ParseArgs()

	var rootDb *mymysql.MySQL
	// var leafDb *mymysql.MySQL

	// verify database connectability
	if *rootHost != "" {
		rootAddr := fmt.Sprintf("%v:%v", *rootHost, *rootPort)
		rootDb = mymysql.New("tcp", "", rootAddr, *rootUser, *rootPass)
		rootDb.Register("set names utf8")
		rows, _, err := rootDb.QueryAC("SELECT * FROM mysql.user")
		if err != nil {
			log.Error(err)
			return
		}
		// defer log.Info(rootDb)
		// defer rootDb.Close()
		// defer log.Info(rootDb)
		log.Info(len(rows))
		for i, _ := range rows {
			log.Info("%v\n", i)
		}
		log.Close()
	}
}
