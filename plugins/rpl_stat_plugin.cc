/**
   Replication plugin for MySQL.
*/

#include <mysql/plugin.h>
#include "sql_class.h"
#include <replication.h>
#include <hash.h>

#define LOG_LINE_BUFFER_SIZE 150

static HASH binlog_states;
static mysql_mutex_t LOCK_rpl_log; /* rpl log file mutex */
static mysql_mutex_t LOCK_hash;    /* HASH struct write lock */
static File rpl_stat_log_file;     /* the log file to write to */


/**
   HASH helpers
*/
struct st_rpl_stat_element
{
  char *server_id;              /* range: 0-4294967295 */
  char *log_file;               /* the binlog file in read */
  my_off_t log_pos;             /* the log position inside binlog file */
};
uchar* get_table_key(const uchar *ptr, size_t *plen, my_bool first)
{
  st_rpl_stat_element *element= (st_rpl_stat_element *)ptr;
  *plen= strlen(element->server_id);
  return (uchar *) element->server_id;
}
void free_element(void *data)
{
  st_rpl_stat_element *element= (st_rpl_stat_element *)data;
  
  my_free(element->server_id);
  my_free(element->log_file);
  my_free(element);
  
  return;
}


/**
   Write log
*/
static int write_stat_log(Binlog_transmit_param *param,
                          const char *log_file, my_off_t log_pos)
{
  char line_buf[LOG_LINE_BUFFER_SIZE];
  char server_id[11];           /* for number in range: 0-4294967295 */

  my_snprintf(server_id, sizeof(server_id), "%d", param->server_id);
  
  /* check previous binlog stat */
  uchar *data= my_hash_search(&binlog_states, (uchar *)server_id,
                              strlen(server_id));
  if (data)                     /* modify old element */
  {
    st_rpl_stat_element *element= (st_rpl_stat_element *)data;
    if (strcmp(element->log_file, log_file) == 0)
    {
      DBUG_RETURN(0);
    }
    else
    {
      my_free(element->server_id);
      my_free(element->log_file);
      element->server_id= my_strdup(server_id, MYF(0));
      element->log_file= my_strdup(log_file, MYF(0));
      element->log_pos= log_pos;
    }
  }
  else                          /* create new element */
  {
    st_rpl_stat_element *element= (st_rpl_stat_element *)
      my_malloc(sizeof(struct st_rpl_stat_element), MYF(0));
    element->server_id= my_strdup(server_id, MYF(0));
    element->log_file= my_strdup(log_file, MYF(0));
    element->log_pos= log_pos;
    mysql_mutex_lock(&LOCK_hash);
    my_hash_insert(&binlog_states, (const uchar *)element);
    mysql_mutex_unlock(&LOCK_hash);
  }
  
  my_snprintf(line_buf, sizeof(line_buf),
              "server_id:%s, binlog_file:%s, offset:%d\n",
              server_id, log_file, log_pos);
  mysql_mutex_lock(&LOCK_rpl_log);
  my_write(rpl_stat_log_file, (uchar*) line_buf,
           strlen(line_buf), MYF(0));
  mysql_mutex_unlock(&LOCK_rpl_log);
  
  DBUG_RETURN(0);
}

/**
   when replication start, write replication position to log immediately
*/
int rpl_stat_transmit_start(Binlog_transmit_param *param,
                            const char *log_file, my_off_t log_pos)
{
  DBUG_ENTER("rpl_stat_transmit_start");
  DBUG_RETURN(write_stat_log(param, log_file, log_pos));
}

/**
   when replicating, write replication position to log every 1000 times
*/
int rpl_stat_before_send_event(Binlog_transmit_param *param,
                               unsigned char *packet, unsigned long len,
                               const char *log_file, my_off_t log_pos)
{
  DBUG_ENTER("rpl_stat_before_send_event");
  DBUG_RETURN(write_stat_log(param, log_file, log_pos));
}


/**
 * build transmit observer structure
 */
Binlog_transmit_observer transmit_observer = {
  sizeof(Binlog_transmit_observer), // len

  rpl_stat_transmit_start,          // start
  NULL,                             // stop
  NULL,                             // reserve_header
  rpl_stat_before_send_event,       // before_send_event
  NULL,                             // after_send_event
  NULL,                             // reset
};

/**
   plugin init function
*/
static int rpl_stat_plugin_init(void *p)
{
  DBUG_ENTER("rpl_stat_plugin_init");
  char log_filename[FN_REFLEN];
  struct st_plugin_int *plugin= (struct st_plugin_int *)p;

  my_hash_init(&binlog_states, &my_charset_latin1, 4, 0, 0,
               (my_hash_get_key) get_table_key, free_element, 0);
  fn_format(log_filename, "rpl-stat", "", ".log",
            MY_REPLACE_EXT | MY_UNPACK_FILENAME);
  mysql_mutex_init(0, &LOCK_rpl_log, MY_MUTEX_INIT_SLOW);
  mysql_mutex_init(0, &LOCK_hash, MY_MUTEX_INIT_SLOW);

  mysql_mutex_lock(&LOCK_rpl_log);
  rpl_stat_log_file= my_open(log_filename, O_RDWR|O_APPEND|O_CREAT, MYF(0));
  if (register_binlog_transmit_observer(&transmit_observer, p))
  {
    // registration failed
    my_close(rpl_stat_log_file, MYF(0));
    mysql_mutex_unlock(&LOCK_rpl_log);
    mysql_mutex_destroy(&LOCK_rpl_log);
    mysql_mutex_destroy(&LOCK_hash);
    DBUG_RETURN(1);
  }
  mysql_mutex_unlock(&LOCK_rpl_log);

  sql_print_information("*plugin %s regisitered", plugin->name.str);
  DBUG_RETURN(0);
}

/**
   plugin de-init function
*/
static int rpl_stat_plugin_deinit(void *p)
{
  DBUG_ENTER("rpl_stat_plugin_deinit");

  mysql_mutex_lock(&LOCK_rpl_log);
  my_close(rpl_stat_log_file, MYF(0));
  mysql_mutex_unlock(&LOCK_rpl_log);
  mysql_mutex_destroy(&LOCK_rpl_log);
  mysql_mutex_destroy(&LOCK_hash);
  
  if (unregister_binlog_transmit_observer(&transmit_observer, p))
  {
    sql_print_error("unregister_binlog_transmit_observer failed");
    DBUG_RETURN(1);
  }
  
  DBUG_RETURN(0);
}


/**
   build plugin requisites
*/
struct Mysql_replication rpl_stat_plugin=
{ MYSQL_REPLICATION_INTERFACE_VERSION };


/**
  plugin library descriptor
*/
mysql_declare_plugin(rpl_stat_plugin)
{
  MYSQL_REPLICATION_PLUGIN,
  &rpl_stat_plugin,
  "rpl_stat_plugin",
  "Vincent Gu (lisnaz@gmail.com)",
  "A plugin which will write replication progress info to log file on master",
  PLUGIN_LICENSE_GPL,
  rpl_stat_plugin_init,
  rpl_stat_plugin_deinit,
  0x0100,
  NULL,
  NULL,
  NULL
}
mysql_declare_plugin_end;
