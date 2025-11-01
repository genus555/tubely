package main

import (
	"encoding/base64"
	"crypto/rand"
	"os/exec"
	"bytes"
	"math"
)

func MakeFileName() string {
	key := make([]byte, 32)
	rand.Read(key)
	file_name := base64.RawURLEncoding.EncodeToString(key)

	return file_name
}

func getVideoAspectRatio(filePath string) (string ,error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var b bytes.Buffer
	cmd.Stdout = &b
	if err := cmd.Run(); err != nil {return "", err}
	result, err := unMarshal(b.Bytes())
	if err != nil {return "", err}

	ratio := float64(result.Streams[0].Width)/float64(result.Streams[0].Height)
	if math.Abs(ratio - (16.0/9.0)) < 0.03 {
		return "16:9", nil
	} else if math.Abs(ratio - (9.0/16.0)) < 0.03 {
		return "9:16", nil
	} else {
		return "other", nil
	}
}

func aspectRatioType(ratio string) (string) {
	switch ratio {
	case "16:9":
		return "landscape"
	case "9:16":
		return "portrait"
	default:
		return "other"
	}
}