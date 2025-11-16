package main

import (
	"strings"
	"time"
	"fmt"

	"github.com/genus555/tubely/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	raw_vid_url := *video.VideoURL
	if strings.HasPrefix(raw_vid_url, "http") {return video, nil}
	bucket_key := strings.Split(raw_vid_url, ",")
	if len(bucket_key) != 2 {
		return video, fmt.Errorf("invalid video_url: %q", raw_vid_url)
	}
	bucket := strings.TrimSpace(bucket_key[0])
	key := strings.TrimSpace(bucket_key[1])
	duration := time.Duration(1 * time.Minute)
	if bucket == "" || key == "" {
		return video, nil
	}

	presigned_url, err := generatePresignedURL(cfg.s3Client, bucket, key, duration)
	if err != nil {return video, err}

	video.VideoURL = &presigned_url

	return video, nil
}
