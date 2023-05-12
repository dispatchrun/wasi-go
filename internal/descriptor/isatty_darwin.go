package descriptor

import "golang.org/x/sys/unix"

// IsATTY is true if the file descriptor fd refers to a valid terminal type device.
func IsATTY(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	return err == nil
}
