package main

// #include <termios.h>
// #include <stdlib.h>
// #include <stdio.h>
// 
// void do_print() {
//   printf("hi!\n");
// }
import "C"

func main() {
	C.do_print()


}
