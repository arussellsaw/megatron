package subdb

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/martinlindhe/subtitles"
	"github.com/pkg/errors"
)

func GetSubs(path string) (*subtitles.Subtitle, error) {
	subPath := path[:len(path)-3] + "en.srt"
	f, err := os.Open(subPath)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	st, err := subtitles.Parse(buf)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func getSRT(path string) ([]byte, error) {
	hash, err := getHash(path)
	if err != nil {
		return nil, errors.Wrap(err, "getting file hash")
	}

	req, err := http.NewRequest("GET", "http://api.thesubdb.com/?action=download&hash="+hash+"&language=en", nil)
	if err != nil {
		return nil, errors.Wrap(err, "constructing hash")
	}
	req.Header.Add("User-Agent", "SubDB/1.0 (TheMegatron/1.0; http://github.com/arussellsaw/megatron)")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Non-200 status code recieved: %s", res.Status)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading body")
	}

	return buf, nil
}

func getHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer f.Close()

	h := md5.New()
	buf := make([]byte, 64*1024)
	if _, err = f.Read(buf); err != nil {
		return "", errors.Wrap(err, "reading initial 64k")
	}

	if _, err := h.Write(buf); err != nil {
		return "", errors.Wrap(err, "writing hash buffer")
	}

	if _, err = f.Seek(-64*1024, 2); err != nil {
		return "", errors.Wrap(err, "seeking to end")
	}

	buf = make([]byte, 64*1024)
	if _, err = f.Read(buf); err != nil {
		return "", errors.Wrap(err, "reading final 64k")
	}

	if _, err := h.Write(buf); err != nil {
		return "", errors.Wrap(err, "writing final hash buffer")
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
