// mtc-cordump will make a still image of a leaf MySQL instance with
// point-in-time informations regarding to all preceding instances.
// 
// For example, using mtc-cordump to make an still image of a random MySQL
// instance N3, which itself is inside a replication chain:
//
//     N1 <- N2 <- N3 <- N4   (N3 replicate from N2, etc...)
//
// Then with the login informations of N1 and N3 and some proper privilege
// setups on them, mtc-cordump can obtain the dump with point-in-time binlog
// positions of every preceding instances, like:
// 
//     N1: mysql-bin.000005 48471238 "2011-09-12 12:32:49"
//     N2: mysql-bin.000023 12389772 "2011-09-12 19:21:43"
//     N3: mysql-bin.000001 21383457 "2011-09-12 08:11:53"
//
// In which case, one can restore this backup with log sequences respective to
// each server in chain.
package main

import (
	"os"
	"flag"
	"fmt"
	"strings"
	l4g "log4go.googlecode.com/hg"
	// mymysql "github.com/ziutek/mymysql"
)

var (
	// args:
	cmdname string        = os.Args[0]
	fs      *flag.FlagSet = flag.NewFlagSet(cmdname, flag.ExitOnError)
	// args: general
	verbose   *bool = fs.Bool("v", false, "verbose output")
	debugMode *bool = fs.Bool("debug", false, "debug mode")
	// args: dumping elements
	dumpDb     *string = fs.String("d", "", "\"db1,db2,...\", include only these db")
	dumpDbExc  *string = fs.String("D", "", "\"db1,db2,...\", exclude these db")
	dumpTb     *string = fs.String("t", "", "\"db1.tb1,db1.tb2,...\", include only these tables")
	dumpTbExc  *string = fs.String("T", "", "\"db1.tb1,db1.tb2,...\", exclude these tables")
	dumpHeight *int    = fs.Int("height", 0, "dump height, default value includes all upstream nodes")
	nodes  []Node = make([]Node, 0, 10)
	// logging controls
	log = make(l4g.Logger)
)

type Node struct {
	host string
	port int
	user string
	pass string
	masterHost string
	masterFile string
	masterPos int64
}

func ParseArgs() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Arg parsing failed: %v\n", err)
		}
	}()

	fs.Usage = func() {
		fmt.Printf("Usage of %v:\n", cmdname)
		fmt.Printf("  %v [OPTIONS] NID [NID...] > backup_file.sql\n\n",
			cmdname)
		fmt.Printf("\nOPTION:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[1:])
	if fs.NArg() == 0 {
		log.Error("wrong args")
		fs.Usage()
		os.Exit(1)
	}
	// parse nodes
	dumpNodes := fs.Args()
	for _, node := range dumpNodes {
		params := strings.Split(node, ",")
		for _, param := range params {
			var host, user, pass string
			var masterHost, masterFile string
			tokens := strings.Split(param, "=")
			if len(tokens) < 2 || len(tokens[1]) == 0 {
				err := fmt.Sprintf("bad formatted Nid: \"%v\"\n", node)
				panic(err)
			}
			switch tokens[0] {
			case "h":
				host = tokens[1]
			case "P":
				port = tokens[1]
			}
			
		}
		// append(nodes, Node{host: })
	}
}

func main() {
	log.AddFilter("stderr", l4g.DEBUG,
		l4g.NewFormatLogWriter(os.Stderr, "[%d %t] [%L] %M"))
	defer log.Close()
	ParseArgs()

	// var rootDb *mymysql.MySQL
	// var leafDb *mymysql.MySQL

	// verify database connectability
	// if *rootHost != "" {
	// 	rootAddr := fmt.Sprintf("%v:%v", *rootHost, *rootPort)
	// 	rootDb = mymysql.New("tcp", "", rootAddr, *rootUser, *rootPass)
	// 	rootDb.Register("set names utf8")
	// 	rows, _, err := rootDb.QueryAC("SELECT * FROM mysql.user")
	// 	if err != nil {
	// 		log.Error(err)
	// 		return
	// 	}
	// 	// defer log.Info(rootDb)
	// 	// defer rootDb.Close()
	// 	// defer log.Info(rootDb)
	// 	log.Info(len(rows))
	// 	for i, _ := range rows {
	// 		log.Info("%v\n", i)
	// 	}
	// 	log.Close()
	// }
}
