/* 
 * Replication plugin for MySQL.
 */




struct st_mysql_information_schema is_repl_stat
{ MYSQL_INFORMATION_SCHEMA_INTERFACE_VERSION };

/*
 * plugin library descriptor
 */
mysql_declare_plugin(is_repl_stat_plugin)
{
  MYSQL_INFORMATION_SCHEMA_PLUGIN,
  &is_rpl_stat,
  "replication_progress_plugin",
  "Vincent Gu (lisnaz@gmail.com)",
  "A plugin which will send replication progress info to I_S table",
  PLUGIN_LICENSE_GPL,
  is_rpl_stat_plugin_init,
  is_rpl_stat_plugin_deinit,
  0x0100,
  NULL,
  NULL,
  NULL
}
mysql_declare_plugin_end;
