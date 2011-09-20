/*
 mysql_backup do the dirty job for you.
*/
package main

import (
	"flag"
	"log"
	mymysql "github.com/ziutek/mymysql"
)

var (
	rootHost *string = flag.String("rh", "host", "host of upstream topmost MySQL instance")
	rootPort *int    = flag.Int("rP", 3306, "port of upstream topmost MySQL instance")
	rootUser *string = flag.String("ru", "rman", "username of upstream topmost MySQL instance")
	rootPass *string = flag.String("rp", "", "password of upstream topmost MySQL instance")

	leafHost *string = flag.String("lh", "host", "host of target MySQL instance")
	leafPort *int = flag.Int("lP", "")
	leafUser = "rman"
	leafPass = ""
)

func init() {
	// prepare command-line flags
	flag.Var(&rootHost, "rh", "host of topmost instance of replication chain")
	flag.Var(&rootPort, "rp", "port of topmost instance of replication chain")
	flag.Var(&rootUser, "ru", "user of topmost instance of replication chain")
	flag.Var(&rootPass, "rp", "passwd of topmost instance of replication chain")
}

func main() {
	flag.Parse()

}
