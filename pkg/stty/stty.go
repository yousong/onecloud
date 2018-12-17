package stty

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

//
//	static void
//	display_recoverable (struct termios *mode)
//	{
//	  printf ("%lx:%lx:%lx:%lx",
//	          (unsigned long int) mode->c_iflag,
//	          (unsigned long int) mode->c_oflag,
//	          (unsigned long int) mode->c_cflag,
//	          (unsigned long int) mode->c_lflag);
//	  for (size_t i = 0; i < NCCS; ++i)
//	    printf (":%lx", (unsigned long int) mode->c_cc[i]);
//	  putchar ('\n');
//	}
//

const (
	NCCS = 19

	raw    = "0:4:bf:8a38:3:1c:7f:15:4:0:1:0:11:13:1a:0:12:f:17:16:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0"
	cooked = "526:5:bf:8a3b:3:1c:7f:15:4:0:1:0:11:13:1a:0:12:f:17:16:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0"
	sane   = "2526:5:bf:8a3b:3:1c:7f:15:4:0:1:0:11:13:1a:0:12:f:17:16:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0"

	bash = "4140:5:bf:8a31:3:0:7f:15:4:0:1:0:0:0:0:ff:12:0:17:0:ff:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0"
	zsh  = "4400:5:bf:8a31:3:1c:7f:15:4:0:1:0:11:13:1a:ff:12:f:17:0:ff:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0"
)

var (
	TermiosRaw    *unix.Termios
	TermiosCooked *unix.Termios
	TermiosSane   *unix.Termios

	TermiosBash *unix.Termios
	TermiosZsh  *unix.Termios
)

func init() {
	TermiosRaw = MustParseSttyRecoverable(raw)
	TermiosCooked = MustParseSttyRecoverable(cooked)
	TermiosSane = MustParseSttyRecoverable(sane)
	TermiosBash = MustParseSttyRecoverable(bash)
	TermiosZsh = MustParseSttyRecoverable(zsh)
}

func MustParseSttyRecoverable(s string) *unix.Termios {
	termios, err := ParseSttyRecoverable(s)
	if err != nil {
		panic("parse stty recoverable " + s + ": " + err.Error())
	}
	return termios
}

func ParseSttyRecoverable(s string) (*unix.Termios, error) {
	min := NCCS + 3
	cs := strings.Split(s, ":")
	if len(cs) < min {
		return nil, fmt.Errorf("expecting at least %d hexdecimal num, got %d", min, len(cs))
	}
	termios := &unix.Termios{}
	readF := func(msgFmt, hex string, a interface{}) error {
		n, err := fmt.Sscanf(hex, "%x", a)
		if n != 1 || err != nil {
			return fmt.Errorf(msgFmt, hex, err)
		}
		return nil
	}
	if err := readF("parse iflag %s: %s", cs[0], &termios.Iflag); err != nil {
		return nil, err
	}
	if err := readF("parse oflag %s: %s", cs[1], &termios.Oflag); err != nil {
		return nil, err
	}
	if err := readF("parse cflag %s: %s", cs[2], &termios.Cflag); err != nil {
		return nil, err
	}
	if err := readF("parse lflag %s: %s", cs[3], &termios.Lflag); err != nil {
		return nil, err
	}
	ccs := cs[4:]
	for i := 0; i < NCCS; i++ {
		msgFmt := fmt.Sprintf("parse cc[%d]: %%s: %%s", i)
		err := readF(msgFmt, ccs[i], &termios.Cc[i])
		if err != nil {
			return nil, err
		}
	}
	return termios, nil
}

func GetTermios(fd int) (*unix.Termios, error) {
	return unix.IoctlGetTermios(fd, unix.TCGETS)
}

func SetTermios(fd int, termios *unix.Termios) error {
	return unix.IoctlSetTermios(fd, unix.TCSETS, termios)
}

func SetTermiosEx(fd int, termios *unix.Termios) (*unix.Termios, error) {
	oldTermios, err := GetTermios(fd)
	if err != nil {
		return nil, err
	}
	if err := SetTermios(fd, termios); err != nil {
		return nil, err
	}
	return oldTermios, nil
}
