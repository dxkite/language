// test escape
#if 'A' == '\301'
# define SQLITE_EBCDIC 1
#else
# define SQLITE_ASCII 1
#endif