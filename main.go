package main

import (
	"bufio"
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

	if err := os.NewSyscallError("SYS_IOCTL", fmt.Errorf("%v", errno)); err != nil {
		return err
	}

	if r1 != 0 {
		return fmt.Errorf("Error")
	}

	return nil
}

func tty_raw() error {
	raw := orig_termios

	raw.Iflag &= ^(BRKINT | ICRNL | INPCK | ISTRIP | IXON)
	raw.Oflag &= ^(OPOST)
	raw.Cflag |= (CS8)
	raw.Lflag &= ^(ECHO | ICANON | IEXTEN | ISIG)

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
		up            []byte = []byte("\033[A")
		eightbitchars [256]byte
	)

	for i := range eightbitchars {
		eightbitchars[i] = byte(i)
	}

	for {
		bytesread, errno = syscall.Read(ttyfd, c_in[0:])
		if err = os.NewSyscallError("SYS_READ", fmt.Errorf("%v", errno)); err != nil {
			return
		} else if bytesread < 0 {
			return fmt.Errorf("read error")
		}

		if bytesread == 0 {
			c_out[0] = 'T'
			_, errno = syscall.Write(ttyfd, c_out[0:])
			if err = os.NewSyscallError("SYS_WRITE", fmt.Errorf("%v", errno)); err != nil {
				return
			}
		} else {
			switch c_in[0] {
			case 'q':
				return nil
			case 'z':
				_, errno = syscall.Write(ttyfd, []byte{'Z'})
				if err = os.NewSyscallError("SYS_WRITE", fmt.Errorf("%v", errno)); err != nil {
					return nil
				}
			case 'u':
				_, errno = syscall.Write(ttyfd, up)
				if err = os.NewSyscallError("SYS_WRITE", fmt.Errorf("%v", errno)); err != nil {
					return nil
				}
			default:
				c_out[0] = '*'
				_, errno = syscall.Write(ttyfd, c_out[0:])
				if err = os.NewSyscallError("SYS_WRITE", fmt.Errorf("%v", errno)); err != nil {
					return nil
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

	if err = tty_raw(); err != nil {
		fmt.Println(err)
		return
	}
	// try reading a line
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		setTermios(&orig_termios)
		fmt.Println("ERR:", err)
	}
	err = setTermios(&orig_termios)
	if err != nil {
		fmt.Println("ERR2:", err)
	}
	fmt.Println("GOT:", string(line))

	return
}
