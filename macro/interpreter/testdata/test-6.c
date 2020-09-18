#define READ_UTF8(c)  "READ_UTF8("c")"
READ_UTF8(c);
#define WINDOWFUNCALL(name,nArg,extra) {                                   \
  nArg, (SQLITE_UTF8|SQLITE_FUNC_WINDOW|extra), 0, 0,                      \
  name ## StepFunc, name ## FinalizeFunc, name ## ValueFunc,               \
  name ## InvFunc, name ## Name, {0}                                       \
}
WINDOWFUNCALL(percent_rank, 0, 0)
# define IOTRACE(A)  if( sqlite3IoTrace ){ sqlite3IoTrace A; }
IOTRACE(("UNLOCK %p %d\n", pPager, eLock))
#define put32bits(A,B)  sqlite3Put4byte((u8*)A,B)
put32bits(((char*)pPg->pData)+24, change_counter);