package editor

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
)

func Resolve() (string, error) {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if v := os.Getenv(env); v != "" {
			slog.Debug("editor from env", "var", env, "bin", v)
			return v, nil
		}
	}
	if path, err := exec.LookPath("vi"); err == nil {
		slog.Debug("editor fallback", "bin", path)
		return path, nil
	}
	return "", errors.New("no editor found (set $VISUAL or $EDITOR)")
}

func Run(path string) error {
	bin, err := Resolve()
	if err != nil {
		return err
	}
	slog.Debug("launching editor", "bin", bin, "file", path)

	cmd := exec.Command(bin, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
