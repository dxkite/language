#define A "[A]"
#define B(b) B"{{"b"}}"
#define C(c) #c B(c) line(__LINE__)
#define E(x,y) #x #y A B(y) y A
#define F(x,y) x##y
#define G(x,y) x,y
C(10086)
CC(10086, dxkite E(12 34 , 5 ))
F(A,A)
F(123 C(5) C, B(1) A)
G(C(5), A)