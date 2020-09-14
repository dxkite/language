#define ABC "<ABC>"
#define EEE(x) "<[[" x "]]>"
#define BB(b) #b EEE(b) "line<"__LINE__">"
#define y  "<YYYY>"
#define E(x,y) #x #y ABC BB(y) y
#define EE(x,y) x##y x

BB (10086)
CC (10086,  112 233 E(12 33, 5 ) )
E(12 33, ABC)
EE(ABC,DEF)