package tty

import (
	"io"
	"os"
)

func ReadStdinIfPiped(stdin *os.File) (string, error) {
	info, err := stdin.Stat()
	if err != nil {
		return "", err
	}
	if (info.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
