#define A +10
#define B(x) x+10+__LINE__
#define line(x) <line:x>
#if B(10 A) == 30
"B(10 A) is 30" line(__LINE__)
#else
"B(10 A) is not 30" line(__LINE__)
#endif
#define V __LINE__
#if V > 10
line > 10 line(__LINE__)
#else
line < 10 line(__LINE__)
#endif
"V is" V
#if 'a' << 1.1
'a' << 1.1 line(__LINE__)
#endif
#if defined(A) + defined(B) >1
A and B is defined line(__LINE__)
#endif
line(__LINE__)
#define AA 100
#if AA < 10
AA < 10 line(__LINE__)
#elif AA > 100
AA > 100 line(__LINE__)
#else
AA = 100 line(__LINE__)
#endif
line(__LINE__)
#undef line
line(__LINE__)
__FILE__ __LINE__
B(100)