#!/usr/bin/env perl
#
# A tool which purge binlogs up to most recent file.
#
# MySQL master needs a USER with SUPER privilege, slave needs REPLICATION CLIENT
# privilege.

use strict;
use warnings;

use DBI;

# CONFIG VARIABLES
my $m_hostname = "localhost";
my $m_db_name = "test";
my $m_db_port = 3306;
my $s_hostname = 'mysql4';
my $s_db_name = 'test';
my $s_db_port = 3306;
my $m_dsn = "dbi:mysql:db=$m_db_name;$m_hostname:$m_db_port";
my $s_dsn = "dbi:mysql:db=$s_db_name;host=$s_hostname;port=$s_db_port";
my $user = "binlog";
my $pw = "binlog";


sub main {
    # inquiry slave's log position
    my $s_dbh = DBI->connect($s_dsn, $user, $pw)
        or die DBI::errstr;
    my @ary = $s_dbh->selectrow_array('SHOW SLAVE STATUS')
        or die $s_dbh->errstr;
    my $binlog_filename = $ary[5];

    # purge binlog(s) on master
    my $m_dbh = DBI->connect($m_dsn, $user, $pw)
        or die DBI::errstr;
    my $rv = $m_dbh->do("PURGE BINARY LOGS TO '$binlog_filename'")
        or die m_dbh->errstr;
    if ($rv > 0) {
        die "Error while purging binlog(s) on master!";
    }

    return $rv;
}

main();
