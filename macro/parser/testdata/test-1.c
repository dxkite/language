#ifndef __MY_HEAD__
# define __MY_HEAD__
int main();
#include <stdio.h>
# define PRINT print("hello world")
# if C_VERSION > 2000L
int main() {
	PRINT;
}
# elif defined(_MSC_VER) || defined(__BORLANDC__)
int main(int argc, char **argv) {
	PRINT;
}
#else
# error hello world
# endif
#endif // __MY_HEAD__
#pragma warning(disable : 4054)
#pragma warning(disable : 4055)
#pragma warning(disable : 4100)