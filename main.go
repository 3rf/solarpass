package main

import (
	//	"bufio"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

// termios types
type cc_t byte
type speed_t uint32
type tcflag_t uint32

// termios constants
const (
	SYS_IOCTL = 54
)

// ioctl constants
const (
	TCGETS = 21517
	TCSETS = 21518
)

var (
	orig_termios syscall.Termios
	ttyfd        int = 0 // STDIN_FILENO
)

func getTermios(dst *syscall.Termios) error {
	r1, _, errno := syscall.Syscall(SYS_IOCTL,
		uintptr(ttyfd), uintptr(TCGETS),
		uintptr(unsafe.Pointer(dst)))

	if errno != 0 {
		if err := os.NewSyscallError("SYS_IOCTL", fmt.Errorf("%v", errno)); err != nil {
			return err
		}

		if r1 != 0 {
			return fmt.Errorf("Error")
		}
	}

	return nil
}

func setTermios(src *syscall.Termios) error {
	r1, _, errno := syscall.Syscall(SYS_IOCTL,
		uintptr(ttyfd), uintptr(TCSETS),
		uintptr(unsafe.Pointer(src)))

	if errno != 0 {
		return os.NewSyscallError("SYS_IOCTL", errno)
	}
	if r1 != 0 {
		return fmt.Errorf("Error")
	}

	return nil
}

func tty_hidden() ([]byte, error) {
	newState := orig_termios
	newState.Lflag &^= syscall.ECHO
	//newState.Lflag |= syscall.ICANON | syscall.ISIG
	//newState.Iflag |= syscall.ICRNL
	if err := setTermios(&newState); err != nil {
		return nil, err
	}
	defer setTermios(&orig_termios)

	var buf [16]byte
	var ret []byte
	for {
		n, err := syscall.Read(ttyfd, buf[:])
		if err != nil {
			return nil, err
		}
		if n == 0 {
			if len(ret) == 0 {
				return nil, io.EOF
			}
			break
		}
		if buf[n-1] == '\n' {
			n--
		}
		ret = append(ret, buf[:n]...)
		if n < len(buf) {
			break
		}
	}

	return ret, nil
}

func main() {
	var (
		err error
	)

	if err = getTermios(&orig_termios); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("GOT TERMIOS!")

	fmt.Printf("%#v\n\n", orig_termios)

	fmt.Printf("Passwd plz:")

	read, err := tty_hidden()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\nGOT:", string(read))

	return
}
