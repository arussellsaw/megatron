package ffmpeg

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"time"
)

func Gif(path, text string, start, end time.Duration) (*os.File, error) {
	id := hash(path, text, start, end)
	f, err := os.Create(id + ".txt")
	if err != nil {
		return nil, err
	}
	f.Write([]byte(text))
	f.Close()
	dur := end - start
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-ss", fmt.Sprintf("%f", start.Seconds()),
		"-t", fmt.Sprintf("%f", dur.Seconds()),
		"-i", path,
		"-vf", `drawtext=fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf: textfile=`+id+`.txt: fontcolor=white: fontsize=24: box=1: boxcolor=black@0.5: boxborderw=5: x=(w-text_w)/2: y=300`,
		"-f", "gif", id+".gif",
	)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return os.Open(id + ".gif")
}

func hash(v ...interface{}) string {
	h := fnv.New32()
	h.Write([]byte(fmt.Sprint(v...)))
	return hex.EncodeToString(h.Sum(nil))
}
