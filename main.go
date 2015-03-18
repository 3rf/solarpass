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
	newState.Lflag |= syscall.ICANON | syscall.ISIG
	newState.Iflag |= syscall.ICRNL
	if err := setTermios(&newState); err != nil {
		return nil, err
	}

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

func screenio() (err error) {
	var (
		bytesread     int
		c_in, c_out   [1]byte
		eightbitchars [256]byte
	)

	for i := range eightbitchars {
		eightbitchars[i] = byte(i)
	}

	for {
		bytesread, err = syscall.Read(ttyfd, c_in[0:])
		if bytesread < 0 {
			return fmt.Errorf("read error")
		}

		if bytesread == 0 {
			c_out[0] = 'T'
			_, _ = syscall.Write(ttyfd, c_out[0:])
		} else {
			switch c_in[0] {
			case 'q':
				return nil
			case 'z':
				_, err = syscall.Write(ttyfd, []byte{'Z'})
				if err != nil {
					return err
				}
			default:
				c_out[0] = '*'
				_, err = syscall.Write(ttyfd, c_out[0:])
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
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

	fmt.Println("GOT:", string(read))

	return
}
