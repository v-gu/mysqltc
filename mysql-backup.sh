#!/bin/bash

# Host name (or IP address) of MySQL server e.g localhost
DBHOST=
# Default port number
DBPORT=3306
# Default MySQL cnf file
DBCNF=/etc/mysql/my.cnf
# Sould we lock all MySQL tables while dumping
LOCK='false'
# Separate backup directory and file for each DB? (yes or no)
SEPDIR=no

# Help info.
usage()
{
cat << EOF
usage: $0 OPTIONS

This script will backup remote mysql instance's data&cnfig to local.

OPTIONS:
   -c      MySQL cnf file, default: /etc/mysql/my.cnf
   -h      Show help
   -H      MySQL host
   -l      with lock, only applies to ALL mode, default: no lock
   -P      MySQL port, default: 3306
   -s      Seperate backup directory for each schema, default: no seperation
EOF
}

# check command-line arguments count
while getopts "c:hlH:P:s" OPTION;do
    case $OPTION in
        c)
        DBCNF=$OPTARG
        ;;
    h)
        usage
        exit 1
        ;;
    H)
        DBHOST=$OPTARG
        ;;
    l)
        LOCK='true'
        ;;
    P)
        DBPORT=$OPTARG
        ;;
    s)
        SEPDIR=yes
        ;;
    ?)
        usage
        exit 1
        ;;
    esac
done

if [[ -z $DBHOST ]];then
   usage
   exit 1
fi

# Username to access the MySQL server e.g. dbuser
USERNAME=pwbackup

# Username to access the MySQL server e.g. password
PASSWORD=PwPwBakUp1

# List of DBNAMES for Daily/Weekly Backup e.g. "DB1 DB2 DB3"
DBNAMES="all"

# Backup directory location e.g /backups
BACKUPDIR="/perfectworld/mysql_data/$DBHOST"
if [[ "$DBPORT" != "3306" ]];then
   BACKUPDIR="$BACKUPDIR"-$DBPORT
fi

DESTDIR="/backup/mysql"

DAYTOKEEP=3

# Mail setup
# What would you like to be mailed to you?
# - log   : send only log file
# - files : send log file and sql files as attachments (see docs)
# - stdout : will simply output the log to the screen if run manually.
# - quiet : Only send logs if an error occurs to the MAILADDR.
MAILCONTENT="stdout"

# Set the maximum allowed email size in k. (4000 = approx 5MB email [see docs])
MAXATTSIZE="4000"

# Email Address to send mail to? (user@domain.com)
MAILADDR="don.nguyen@perfectworld.com,stuart.wang@perfectworld.com,guchunjiang@wanmei.com"

# ============================================================
# === ADVANCED OPTIONS ( Read the doc's below for details )===
#=============================================================

# List of DBBNAMES for Monthly Backups.
MDBNAMES="all"

# List of DBNAMES to EXLUCDE if DBNAMES are set to all (must be in " quotes)
DBEXCLUDE="mysql information_schema performance_schema"

# Include CREATE DATABASE in backup?
CREATE_DATABASE=yes

# Which day do you want weekly backups? (1 to 7 where 1 is Monday)
DOWEEKLY=6

# Choose Compression type. (gzip or bzip2)
COMP=gzip

# Compress communications between backup server and MySQL server?
COMMCOMP=no

# Additionally keep a copy of the most recent backup in a seperate directory.
LATEST=no

#  The maximum size of the buffer for client/server communication. e.g. 16MB (maximum is 1GB)
MAX_ALLOWED_PACKET=

#  For connections to localhost. Sometimes the Unix socket file must be specified.
SOCKET=

# Command to run before backups (uncomment to use)
#PREBACKUP="/etc/mysql-backup-pre"

# Command run after backups (uncomment to use)
#POSTBACKUP="/etc/mysql-backup-post"

# === Advanced options doc's ===
#
# The list of MDBNAMES is the DB's to be backed up only monthly. You should
# always include "mysql" in this list to backup your user/password
# information along with any other DBs that you only feel need to
# be backed up monthly. (if using a hosted server then you should
# probably remove "mysql" as your provider will be backing this up)
# NOTE: If DBNAMES="all" then MDBNAMES has no effect as all DBs will be backed
# up anyway.
#
# If you set DBNAMES="all" you can cnfigure the option DBEXCLUDE. Other
# wise this option will not be used.
# This option can be used if you want to backup all dbs, but you want 
# exclude some of them. (eg. a db is to big).
#
# Set CREATE_DATABASE to "yes" (the default) if you want your SQL-Dump to create
# a database with the same name as the original database when restoring.
# Saying "no" here will allow your to specify the database name you want to
# restore your dump into, making a copy of the database by using the dump
# created with automysqlbackup.
# NOTE: Not used if SEPDIR=no
#
# The SEPDIR option allows you to choose to have all DBs backed up to
# a single file (fast restore of entire server in case of crash) or to
# seperate directories for each DB (each DB can be restored seperately
# in case of single DB corruption or loss).
#
# To set the day of the week that you would like the weekly backup to happen
# set the DOWEEKLY setting, this can be a value from 1 to 7 where 1 is Monday,
# The default is 6 which means that weekly backups are done on a Saturday.
#
# COMP is used to choose the copmression used, options are gzip or bzip2.
# bzip2 will produce slightly smaller files but is more processor intensive so
# may take longer to complete.
#
# COMMCOMP is used to enable or diable mysql client to server compression, so
# it is useful to save bandwidth when backing up a remote MySQL server over
# the network. 
#
# LATEST is to store an additional copy of the latest backup to a standard
# location so it can be downloaded bt thrid party scripts.
#
# If the DB's being backed up make use of large BLOB fields then you may need
# to increase the MAX_ALLOWED_PACKET setting, for example 16MB..
#
# When connecting to localhost as the DB server (DBHOST=localhost) sometimes
# the system can have issues locating the socket file.. This can now be set
# using the SOCKET parameter.. An example may be SOCKET=/private/tmp/mysql.sock
#
# Use PREBACKUP and POSTBACKUP to specify Per and Post backup commands
# or scripts to perform tasks either before or after the backup process.
#
#
#=====================================================================
# Backup Rotation..
#=====================================================================
#
# Daily Backups are rotated weekly..
# Weekly Backups are run by default on Saturday Morning when
# cron.daily scripts are run...Can be changed with DOWEEKLY setting..
# Weekly Backups are rotated on a 5 week cycle..
# Monthly Backups are run on the 1st of the month..
# Monthly Backups are NOT rotated automatically...
# It may be a good idea to copy Monthly backups offline or to another
# server..
#
#=====================================================================
# Please Note!!
#=====================================================================
#
# I take no resposibility for any data loss or corruption when using
# this script..
# This script will not help in the event of a hard drive crash. If a 
# copy of the backup has not be stored offline or on another PC..
# You should copy your backups offline regularly for best protection.
#
# Happy backing up...
#
#=====================================================================
# Restoring
#=====================================================================
# Firstly you will need to uncompress the backup file.
# eg.
# gunzip file.gz (or bunzip2 file.bz2)
#
# Next you will need to use the mysql client to restore the DB from the
# sql file.
# eg.
# mysql --user=username --pass=password --host=dbserver database < /path/file.sql
# or
# mysql --user=username --pass=password --host=dbserver -e "source /path/file.sql" database
#
# NOTE: Make sure you use "<" and not ">" in the above command because
# you are piping the file.sql to mysql and not the other way around.
#
# Lets hope you never have to use this.. :)
#
#=====================================================================
#=====================================================================
#
# Should not need to be modified from here down!!
#
#=====================================================================
#=====================================================================
#=====================================================================
PATH=/usr/local/bin:/usr/bin:/bin:/usr/local/mysql/bin 
DATE=`date +%Y-%m-%d_%Hh%Mm`                # Datestamp e.g 2002-09-21
DOW=`date +%A`                            # Day of the week e.g. Monday
DNOW=`date +%u`                        # Day number of the week 1 to 7 where 1 represents Monday
DOM=`date +%d`                            # Date of the Month e.g. 27
M=`date +%B`                            # Month e.g January
W=`date +%V`                            # Week Number e.g 37
VER=2.5                                    # Version Number
LOGFILE=$BACKUPDIR/log/$DBHOST-`date +%N`.log        # Logfile Name
LOGERR=$BACKUPDIR/log/ERRORS_$DBHOST-`date +%N`.log        # Logfile Name
BACKUPFILES=""
OPT="--quote-names --opt"            # OPT string for use with mysqldump ( see man mysqldump )

# Add --compress mysqldump option to $OPT
if [ "$COMMCOMP" = "yes" ];
    then
        OPT="$OPT --compress"
    fi

# Add --compress mysqldump option to $OPT
if [ "$MAX_ALLOWED_PACKET" ];
    then
        OPT="$OPT --max_allowed_packet=$MAX_ALLOWED_PACKET"
    fi

# Create required directories
if [ ! -e "$BACKUPDIR" ]        # Check Backup Directory exists.
    then
    mkdir -p "$BACKUPDIR"
fi

if [ ! -e "$BACKUPDIR/daily" ]        # Check Daily Directory exists.
    then
    mkdir -p "$BACKUPDIR/daily/log"
fi

if [ ! -e "$BACKUPDIR/weekly" ]        # Check Weekly Directory exists.
    then
    mkdir -p "$BACKUPDIR/weekly/log"
fi

if [ ! -e "$BACKUPDIR/monthly" ]    # Check Monthly Directory exists.
    then
    mkdir -p "$BACKUPDIR/monthly/log"
fi

if [ "$LATEST" = "yes" ]
then
    if [ ! -e "$BACKUPDIR/latest" ]    # Check Latest Directory exists.
    then
        mkdir -p "$BACKUPDIR/latest"
    fi
eval rm -fv "$BACKUPDIR/latest/*"
fi

# IO redirection for logging.
touch $LOGFILE
exec 6>&1           # Link file descriptor #6 with stdout.
                    # Saves stdout.
exec > $LOGFILE     # stdout replaced with file $LOGFILE.
touch $LOGERR
exec 7>&2           # Link file descriptor #7 with stderr.
                    # Saves stderr.
exec 2> $LOGERR     # stderr replaced with file $LOGERR.


# Functions

# Database dump function
dbdump () {
mysqldump --user=$USERNAME --password=$PASSWORD --host=$DBHOST --port=$DBPORT $OPT $1 > $2
return 0
}

# Compression function plus latest copy
SUFFIX=""
compression () {
# First argument should be the target file name, following the file names which
# needs to be compressed. If there is only one argument, then compress on that
# file only.
FILENAME="$1"
[ $# != 1 ] && shift
if [ "$COMP" = "gzip" ]; then
    cat "$*" | gzip > "$FILENAME.gz" && rm -f "$FILENAME"
    echo
    echo Backup Information for "$FILENAME"
    gzip -l "$FILENAME.gz"
    SUFFIX=".gz"
elif [ "$COMP" = "bzip2" ]; then
    echo Compression information for "$FILENAME.bz2"
    cat "%*" | bzip2 -f -v -c >"$FILENAME.bz2" 2>&1 && rm -f "$FILENAME"
    SUFFIX=".bz2"
else
    echo "No compression option set, check advanced settings"
fi
if [ "$LATEST" = "yes" ]; then
    cp "$FILENAME$SUFFIX" "$BACKUPDIR/latest/"
fi    
return 0
}


# Run command before we begin
if [ "$PREBACKUP" ]
    then
    echo ======================================================================
    echo "Prebackup command output."
    echo
    eval $PREBACKUP
    echo
    echo ======================================================================
    echo
fi


if [ "$SEPDIR" = "yes" ]; then # Check if CREATE DATABSE should be included in Dump
    if [ "$CREATE_DATABASE" = "no" ]; then
        OPT="$OPT --no-create-db"
    else
        OPT="$OPT --databases"
    fi
else
    OPT="$OPT --databases"
fi

# Hostname for LOG information
if [ "$DBHOST" = "localhost" ]; then
    HOST=`hostname`
    if [ "$SOCKET" ]; then
        OPT="$OPT --socket=$SOCKET"
    fi
else
    HOST=$DBHOST
fi

# If backing up all DBs on the server
if [ "$DBNAMES" = "all" ]; then
        DBNAMES="`mysql --user=$USERNAME --password=$PASSWORD --host=$DBHOST --port=$DBPORT --batch --skip-column-names -e "show databases"| sed 's/ /%/g'`"

    # If DBs are excluded
    for exclude in $DBEXCLUDE
    do
        DBNAMES=`echo $DBNAMES | sed "s/\b$exclude\b//g"`
    done
        MDBNAMES=$DBNAMES
        if [ "$LOCK" == "true" ]; then
            OPT="$OPT --lock-all-tables --master-data=2"
        fi
fi
    
echo ======================================================================
echo Backup of Database Server - $HOST
echo ======================================================================

# Test is seperate DB backups are required
if [ "$SEPDIR" = "yes" ]; then
echo Backup Start Time `date`
echo ======================================================================
    # Monthly Full Backup of all Databases
    if [ $DOM = "01" ]; then
        for MDB in $MDBNAMES
        do
            # Prepare $DB for using
            MDB="`echo $MDB | sed 's/%/ /g'`"
            [ ! -e "$BACKUPDIR/monthly/$MDB" ] && mkdir -p "$BACKUPDIR/monthly/$MDB"
            echo Monthly Backup of $MDB...
            dbdump "$MDB" "$BACKUPDIR/monthly/$MDB/${MDB}_$DATE.$M.$MDB.sql"
            compression "$BACKUPDIR/monthly/$MDB/${MDB}_$DATE.$M.$MDB.sql"
            BACKUPFILES="$BACKUPFILES $BACKUPDIR/monthly/$MDB/${MDB}_$DATE.$M.$MDB.sql$SUFFIX"
            echo ----------------------------------------------------------------------
        done

        # Backup cnf
        echo Backing up cnf file...
        [ ! -d "$BACKUPDIR/monthly/_cnf_" ] && mkdir "$BACKUPDIR/monthly/_cnf_"
        su - mysync -c "scp $DBHOST:$DBCNF /tmp/$DBHOST-$DBPORT.cnf" && \
            cp "/tmp/$DBHOST-$DBPORT.cnf" "$BACKUPDIR/monthly/_cnf_/$DATE-$M.cnf" && \
            echo Backing up cnf file done.
        [ $? != 0 ] && \
            echo Backing up cnf file failed.
        # compress backed up files
        compression "$BACKUPDIR/monthly/_cnf_/$DATE-$M.cnf"
        BACKUPFILES="$BACKUPFILES $BACKUPDIR/monthly/_cnf_/$DATE-$M.cnf$SUFFIX"
    fi

    # Weekly Backup
    if [ $DNOW = $DOWEEKLY ]; then
        for DB in $DBNAMES
        do
            # Prepare $DB for using
            DB="`echo $DB | sed 's/%/ /g'`"
            
            # Create Seperate directory for each DB
            [ ! -e "$BACKUPDIR/weekly/$DB" ] && mkdir -p "$BACKUPDIR/weekly/$DB"

            echo Weekly Backup of Database \( $DB \)
            echo Rotating 5 weeks Backups...
            if [ "$W" -le 05 ];then
                REMW=`expr 48 + $W`
            elif [ "$W" -lt 15 ];then
                REMW=0`expr $W - 5`
            else
                REMW=`expr $W - 5`
            fi
            eval rm -fv "$BACKUPDIR/weekly/$DB_week.$REMW.*" 
            echo
            dbdump "$DB" "$BACKUPDIR/weekly/$DB/${DB}_week.$W.$DATE.sql"
            compression "$BACKUPDIR/weekly/$DB/${DB}_week.$W.$DATE.sql"
            BACKUPFILES="$BACKUPFILES $BACKUPDIR/weekly/$DB/${DB}_week.$W.$DATE.sql$SUFFIX"
        done
        # Backup cnf
        echo Backing up cnf file...
        [ ! -d "$BACKUPDIR/weekly/_cnf_" ] && mkdir "$BACKUPDIR/weekly/_cnf_"
        su - mysync -c "scp $DBHOST:$DBCNF /tmp/$DBHOST-$DBPORT.cnf" && \
            cp "/tmp/$DBHOST-$DBPORT.cnf" "$BACKUPDIR/weekly/_cnf_/$DATE-$W.cnf" && \
            echo Backing up cnf file done.
        [ $? != 0 ] && \
            echo Backing up cnf file failed.
        # compress backed up files
        compression "$BACKUPDIR/weekly/_cnf_/$DATE-$W.cnf"
        BACKUPFILES="$BACKUPFILES $BACKUPDIR/weekly/_cnf_/$DATE-$W.cnf$SUFFIX"
        echo ----------------------------------------------------------------------
    
    # Daily Backup
    else
        for DB in $DBNAMES
        do
            # Prepare $DB for using
            DB="`echo $DB | sed 's/%/ /g'`"
            
            # Create Seperate directory for each DB
            [ ! -e "$BACKUPDIR/daily/$DB" ] && mkdir -p "$BACKUPDIR/daily/$DB"

            echo Daily Backup of Database \( $DB \)
            echo Rotating last weeks Backup...
            eval rm -fv "$BACKUPDIR/daily/$DB/*.$DOW.sql.*" 
            echo
            dbdump "$DB" "$BACKUPDIR/daily/$DB/${DB}_$DATE.$DOW.sql"
            compression "$BACKUPDIR/daily/$DB/${DB}_$DATE.$DOW.sql"
            BACKUPFILES="$BACKUPFILES $BACKUPDIR/daily/$DB/${DB}_$DATE.$DOW.sql$SUFFIX"
        done
        # Backup cnf
        echo Backing up cnf file...
        [ ! -d "$BACKUPDIR/daily/_cnf_" ] && mkdir "$BACKUPDIR/daily/_cnf_"
        su - mysync -c "scp $DBHOST:$DBCNF /tmp/$DBHOST-$DBPORT.cnf" && \
            cp "/tmp/$DBHOST-$DBPORT.cnf" "$BACKUPDIR/daily/_cnf_/$DATE-$DOW.cnf" && \
            echo Backing up cnf file done.
        [ $? != 0 ] && \
            echo Backing up cnf file failed.
        # compress backed up files
        compression "$BACKUPDIR/daily/_cnf_/$DATE-$DOW.cnf"
        BACKUPFILES="$BACKUPFILES $BACKUPDIR/daily/_cnf_/$DATE-$DOW.cnf$SUFFIX"
        echo ----------------------------------------------------------------------
    fi
echo Backup End `date`
echo ======================================================================


else # One backup file for all DBs
echo Backup Start `date`
echo ======================================================================
    # Monthly Full Backup of all Databases
    if [ $DOM = "01" ]; then
        echo Monthly full Backup of \( $MDBNAMES \)...
            dbdump "$MDBNAMES" "$BACKUPDIR/monthly/$DATE.$M.all-databases.sql"
            # Backup cnf
            echo Backing up cnf file...
            [ ! -d "$BACKUPDIR/monthly/_cnf_" ] && mkdir "$BACKUPDIR/monthly/_cnf_"
            su - mysync -c "scp $DBHOST:$DBCNF /tmp/$DBHOST-$DBPORT.cnf" && \
                cp "/tmp/$DBHOST-$DBPORT.cnf" "$BACKUPDIR/monthly/_cnf_/$DATE-$M.cnf" && \
                echo Backing up cnf file done.
            [ $? != 0 ] && \
                echo Backing up cnf file failed.
            # compress backed up files
            compression "$BACKUPDIR/monthly/$DATE.$M.all-databases.sql"
            compression "$BACKUPDIR/monthly/_cnf_/$DATE-$M.cnf"
            BACKUPFILES="$BACKUPFILES $BACKUPDIR/monthly/$DATE.$M.all-databases.sql$SUFFIX $BACKUPDIR/monthly/_cnf_/$DATE-$M.cnf$SUFFIX"
        echo ----------------------------------------------------------------------
    fi

    # Weekly Backup
    if [ $DNOW = $DOWEEKLY ]; then
        echo Weekly Backup of Databases \( $DBNAMES \)
        echo
        echo Rotating 5 weeks Backups...
            if [ "$W" -le 05 ];then
                REMW=`expr 48 + $W`
            elif [ "$W" -lt 15 ];then
                REMW=0`expr $W - 5`
            else
                REMW=`expr $W - 5`
            fi
        eval rm -fv "$BACKUPDIR/weekly/week.$REMW.*" 
        echo
            dbdump "$DBNAMES" "$BACKUPDIR/weekly/week.$W.$DATE.sql"
            # Backup cnf
            echo Backing up cnf file...
            [ ! -d "$BACKUPDIR/weekly/_cnf_" ] && mkdir "$BACKUPDIR/weekly/_cnf_"
            su - mysync -c "scp $DBHOST:$DBCNF /tmp/$DBHOST-$DBPORT.cnf" && \
                cp "/tmp/$DBHOST-$DBPORT.cnf" "$BACKUPDIR/weekly/_cnf_/$DATE-$W.cnf" && \
                echo Backing up cnf file done.
            [ $? != 0 ] && \
                echo Backing up cnf file failed.
            # compress backed up files
            compression "$BACKUPDIR/weekly/week.$W.$DATE.sql"
            compression "$BACKUPDIR/weekly/_cnf_/$DATE-$W.cnf"
            BACKUPFILES="$BACKUPFILES $BACKUPDIR/weekly/week.$W.$DATE.sql$SUFFIX $BACKUPDIR/weekly/_cnf_/$DATE-$W.cnf$SUFFIX"
        echo ----------------------------------------------------------------------
        
    # Daily Backup
    else
        echo Daily Backup of Databases \( $DBNAMES \)
        echo
        echo Rotating last weeks Backup...
        eval rm -fv "$BACKUPDIR/daily/*.$DOW.sql.*" 
        echo
            dbdump "$DBNAMES" "$BACKUPDIR/daily/$DATE.$DOW.sql"
            # Backup cnf
            echo Backing up cnf file...
            [ ! -d "$BACKUPDIR/daily/_cnf_" ] && mkdir "$BACKUPDIR/daily/_cnf_"
            su - mysync -c "scp $DBHOST:$DBCNF /tmp/$DBHOST-$DBPORT.cnf" && \
                cp "/tmp/$DBHOST-$DBPORT.cnf" "$BACKUPDIR/daily/_cnf_/$DATE-$DOW.cnf" && \
                echo Backing up cnf file done.
            [ $? != 0 ] && \
                echo Backing up cnf file failed.
            # compress backed up files
            compression "$BACKUPDIR/daily/$DATE.$DOW.sql"
            compression "$BACKUPDIR/daily/_cnf_/$DATE-$DOW.cnf"
            BACKUPFILES="$BACKUPFILES $BACKUPDIR/daily/$DATE.$DOW.sql$SUFFIX $BACKUPDIR/daily/_cnf_/$DATE-$DOW.cnf$SUFFIX"
        echo ----------------------------------------------------------------------
    fi
echo Backup End Time `date`
echo ======================================================================
fi
echo Total disk space used for backup storage..
echo Size - Location
echo `du -hs "$BACKUPDIR"`
echo
echo ======================================================================

# Run command when we're done
if [ "$POSTBACKUP" ]
    then
    echo ======================================================================
    echo "Postbackup command output."
    echo
    eval $POSTBACKUP
    echo
    echo ======================================================================
fi

#Clean up IO redirection
exec 1>&6 6>&-      # Restore stdout and close file descriptor #6.
exec 1>&7 7>&-      # Restore stdout and close file descriptor #7.

if [ "$MAILCONTENT" = "files" ]
then
    if [ -s "$LOGERR" ]
    then
        # Include error log if is larger than zero.
        BACKUPFILES="$BACKUPFILES $LOGERR"
        ERRORNOTE="WARNING: Error Reported - "
    fi
    #Get backup size
    ATTSIZE=`du -c $BACKUPFILES | grep "[[:digit:][:space:]]total$" |sed s/\s*total//`
    if [ $MAXATTSIZE -ge $ATTSIZE ]
    then
        BACKUPFILES=`echo "$BACKUPFILES" | sed -e "s# # -a #g"`    #enable multiple attachments
        mutt -s "$ERRORNOTE MySQL Backup Log and SQL Files for $HOST - $DATE" $BACKUPFILES $MAILADDR < $LOGFILE        #send via mutt
    else
        cat "$LOGFILE" | mail -s "WARNING! - MySQL Backup exceeds set maximum attachment size on $HOST - $DATE" $MAILADDR
    fi
elif [ "$MAILCONTENT" = "log" ]
then
    cat "$LOGFILE" | mail -s "MySQL Backup Log for $HOST - $DATE" $MAILADDR
    if [ -s "$LOGERR" ]
        then
            cat "$LOGERR" | mail -s "ERRORS REPORTED: MySQL Backup error Log for $HOST - $DATE" $MAILADDR
    fi    
elif [ "$MAILCONTENT" = "quiet" ]
then
    if [ -s "$LOGERR" ]
        then
            cat "$LOGERR" | mail -s "ERRORS REPORTED: MySQL Backup error Log for $HOST - $DATE" $MAILADDR
            cat "$LOGFILE" | mail -s "MySQL Backup Log for $HOST - $DATE" $MAILADDR
    fi
else
    if [ -s "$LOGERR" ]
        then
            cat "$LOGFILE"
            echo
            echo "###### WARNING ######"
            echo "Errors reported during MySQL Backup execution.. Backup failed"
            echo "Error log below.."
            cat "$LOGERR"
    else
        cat "$LOGFILE"
    fi    
fi

if [ -s "$LOGERR" ]
    then
        STATUS=1
    else
        STATUS=0
fi

# Clean up Logfile  -- do not clean up logfile, edited by Vincent
#eval rm -f "$LOGFILE"
#eval rm -f "$LOGERR"

/usr/bin/rsync -avzc $BACKUPDIR $DESTDIR
find $BACKUPDIR -mtime +${DAYTOKEEP} -exec rm -rf {} \;

exit $STATUS

