package util

import (
	"os"
	"os/exec"
	"viedo2m3u8/model"
)

func RunCommand(dir, name string, args ...string) string {

	return string(RunCommandBytes(dir, name, args...))
}

func RunCommandBytes(dir, name string, args ...string) []byte {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	msg, _ := cmd.CombinedOutput()
	return msg
}

func RunCommandWithConsole(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CompressVideo(set *model.Setting, src, desc string) string {
	return RunCommand(set.FfmpegPath, "ffmpeg", "-i", src, "-r", "25", "-c:v", "libx264", "-y", desc)
}
