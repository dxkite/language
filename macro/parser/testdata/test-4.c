#if defined(SQLITE_SYSTEM_MALLOC) \
  + defined(SQLITE_WIN32_MALLOC) \
  + defined(SQLITE_ZERO_MALLOC) \
  + defined(SQLITE_MEMDEBUG)==0
# define SQLITE_SYSTEM_MALLOC 1
#endif

SQLITE_PRIVATE sqlite3_int64 sqlite3StatusValue(int op){
  wsdStatInit;
  assert( op>=0 && op<ArraySize(wsdStat.nowValue) );
  assert( op>=0 && op<ArraySize(statMutex) );
  assert( sqlite3_mutex_held(statMutex[op] ? sqlite3Pcache1Mutex()
                                           : sqlite3MallocMutex()) );
  return wsdStat.nowValue[op];
}