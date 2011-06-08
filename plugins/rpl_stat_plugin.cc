/**
   Replication plugin for MySQL.
*/

//#include <stdlib.h>
#include <my_global.h>
#include <mysql/plugin.h>
#include "sql_class.h"                          // THD
#include <hash.h>
#include <replication.h>

#define LOG_LINE_BUFFER_SIZE 80

static File log_f;

static struct st_rpl_stat_context
{
  unsigned long events_count= 0;
  File rpl_stat_log_file;
} rpl_stat_context;

int rpl_stat_before_send_event(Binlog_transmit_param *param,
                               unsigned char *packet, unsigned long len,
                               const char *log_file, my_off_t log_pos)
{
  static long long counter= 0;
  uint32 server_id= param->server_id;
  char line_buf[LOG_LINE_BUFFER_SIZE];
    
  
  // my_snprintf(line_buf, sizeof(line_buf),
  //             "server_id:%d, binlog_file:%s, offset:%d\n",
  //             server_id, log_file, log_pos);
  my_snprintf(line_buf, sizeof(line_buf),
              "%d\n", counter++);
  extern File log_f;
  my_write(log_f, (uchar*) line_buf, strlen(line_buf), MYF(0));
  
  DBUG_RETURN(0);
}


/**
 * build transmit observer structure
 */
Binlog_transmit_observer transmit_observer = {
  sizeof(Binlog_transmit_observer), // len

  NULL,                             // start
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
  struct rpl_stat_context *context= rpl_stat_context;
  char log_filename[FN_REFLEN];
  struct st_plugin_int *plugin= (struct st_plugin_int *)p;

  context->events_count= 0;
  fn_format(log_filename, "rpl-stat", "", ".log",
            MY_REPLACE_EXT | MY_UNPACK_FILENAME);
  context->rpl_stat_log_file= my_open(log_filename, O_RDWR|O_APPEND|O_CREAT,
                                      MYF(0));
  extern File log_f;
  log_f= context->rpl_stat_log_file;

  if (register_binlog_transmit_observer(&transmit_observer, p))
  {
    // registration failed
    my_close(context->rpl_stat_log_file, MYF(0));
    my_free(context);
    DBUG_RETURN(1);
  } else{
    plugin->data= (void *)context;
  }

  sql_print_information("*plugin %s regisitered", plugin->name.str);
  DBUG_RETURN(0);
}

/**
   plugin de-init function
*/
static int rpl_stat_plugin_deinit(void *p)
{
  DBUG_ENTER("rpl_stat_plugin_deinit");

  struct st_plugin_int *plugin= (struct st_plugin_int *)p;
  struct rpl_stat_context *context= (struct rpl_stat_context *)plugin->data;

  my_close(context->rpl_stat_log_file, MYF(0));

  my_free(context);

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
struct st_mysql_show_var my_status_vars[]= {
  {""},
};

struct Mysql_replication rpl_stat_plugin
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
