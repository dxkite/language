/**
** 递归宏展开
*/
/*递归函数*/
#define A(x) A(x A(x))
A(10)
#undef A
#define A(x) "call(A(x))"(A(x))
A(100)
/*递归变量*/
#define vfsList GLOBAL(sqlite3_vfs *, vfsList)
  for(pVfs = vfsList; pVfs; pVfs=pVfs->pNext){
    if( zVfs==0 ) break;
    if( strcmp(zVfs, pVfs->zName)==0 ) break;
  }