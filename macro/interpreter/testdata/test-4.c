/**
** 递归调用测试
*/
#define A(x) A(x A(x))
A(10)
#undef A
#define A(x) "call(A(x))"(A(x))
A(100)