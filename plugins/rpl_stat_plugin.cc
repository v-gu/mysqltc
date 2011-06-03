/**
   Replication plugin for MySQL.
*/

//#include <stdlib.h>
#include <mysql/plugin.h>
#include "sql_class.h"                          // THD
#include <hash.h>

int is_rpl_stat_before_send_event(Binlog_transmit_param *param,
                                unsigned char *packet, unsigned long len,
                                const char *log_file, my_off_t log_pos)
{
  uint32 server_id = param->server_id;
  sql_print_error=("server_id: %d, log_file: %s", server_id, log_file);
  return 0;
}


/**
 * build transmit observer structure
 */
Binlog_transmit_observer transmit_observer = {
  sizeof(Binlog_transmit_observer), // len

  NULL,                             // start
  NULL,                             // stop
  NULL,                             // reserve_header
  is_rpl_stat_before_send_event,    // before_send_event
  NULL,                             // after_send_event
  NULL,                             // reset
};

static int is_rpl_stat_plugin_init(void *p)
{
  if (register_binlog_transmit_observer(&transmit_observer, p))
    return 1;
  sql_print_error=("*p's name: %s",(*(st_plugin_int) p)->name);
  return 0;
}

static int is_rpl_stat_plugin_deinit(void *p)
{
  return 0;
}


/**
   build plugin requisites
*/
struct Mysql_replication rpl_stat_plugin
{ MYSQL_INFORMATION_SCHEMA_INTERFACE_VERSION };

struct st

/**
  plugin library descriptor
*/
mysql_declare_plugin(repl_stat_plugin)
{
  MYSQL_REPLICATION_PLUGIN,
  &rpl_stat,
  "replication_progress_plugin",
  "Vincent Gu (lisnaz@gmail.com)",
  "A plugin which will send replication progress info to I_S table on master",
  PLUGIN_LICENSE_GPL,
  rpl_stat_plugin_init,
  rpl_stat_plugin_deinit,
  0x0100,
  NULL,
  NULL,
  NULL
}
mysql_declare_plugin_end;
