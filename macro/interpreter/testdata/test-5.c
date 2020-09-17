#define A 10
#define EMPTY
#if ~A > 20
print("^A > 20")
#else
print("^A !> 20")
#endif

#if !defined EMPTY
print("!defined EMPTY");
#else
print("defined EMPTY");
#endif

#if -A > 0
print("-A > 0");
#else
print("-A !> 0");
#endif

#if !A
print("!A");
#else
print("!!A");
#endif
"A"EMPTY"A"