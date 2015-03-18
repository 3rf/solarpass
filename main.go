package main

import (
	//	"bufio"
	"fmt"
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
	BRKINT    = uint32(0000002)
	ICRNL     = uint32(0000400)
	INPCK     = uint32(0000020)
	ISTRIP    = uint32(0000040)
	IXON      = uint32(0002000)
	OPOST     = uint32(0000001)
	CS8       = uint32(0000060)
	ECHO      = uint32(0000010)
	ICANON    = uint32(0000002)
	IEXTEN    = uint32(0100000)
	ISIG      = uint32(0000001)
	VTIME     = uint32(5)
	VMIN      = uint32(6)
	SYS_IOCTL = 54
)

const NCCS = 32

type termios struct {
	Iflag, Oflag, c_cflag, c_lflag tcflag_t
	c_line                         cc_t
	Cc                             [NCCS]cc_t
	c_ispeed, c_ospeed             speed_t
}

// ioctl constants
const (
	TCGETS = 0x5401
	TCSETS = 0x5402
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

func tty_hidden() error {
	raw := orig_termios

	raw.Cflag &= ^(ECHO | ICANON | IEXTEN | ISIG)
	raw.Cc[VMIN] = 1
	raw.Cc[VTIME] = 0

	if err := setTermios(&raw); err != nil {
		return err
	}

	return nil
}

func screenio() (err error) {
	var (
		bytesread     int
		errno         error
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

		fmt.Println("GOTCHAR", c_in)

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

	fmt.Printf("%#v", orig_termios)

	fmt.Printf("Passwd plz:")

	if err = tty_hidden(); err != nil {
		fmt.Println(err)
		return
	}
	// try reading a line
	screenio()
	/*line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		setTermios(&orig_termios)
		fmt.Println("ERR:", err)
	}
	err = setTermios(&orig_termios)
	if err != nil {
		fmt.Println("ERR2:", err)
	}
	fmt.Println("GOT:", string(line))*/

	return
}
