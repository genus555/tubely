package main

import (
	"net/http"
	"mime"
	"os"
	"io"
	"fmt"

	"github.com/genus555/tubely/internal/auth"
	"github.com/genus555/tubely/internal/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	//set upload limit to 1GB
	maxBytes := int64(1 << 30)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	//get videoID from URL path
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	//authenticate user and userID
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	//get video metadata
	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	//authenticate user
	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User is not video owner", err)
		return
	}

	//parse uploaded video file
	if err = r.ParseMultipartForm(32 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, "File is too large", err)
		return
	}
	file, header , err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	//ensure media is video/mp4
	content_type := header.Header.Get("Content-Type")
	if content_type == "" {
		respondWithError(w, http.StatusBadRequest, "Header is empty", nil)
		return
	}
	mediaType, _, err := mime.ParseMediaType(content_type)
	if err != nil {
		respondWithError(w, 400, "Missing or invalid file type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Video is not an mp4", nil)
		return
	}

	//save upload to a temp file
	tmp, err := os.CreateTemp("", "tubely-upload")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create temp file", err)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err = io.Copy(tmp, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to copy upload to temp file", err)
		return
	}

	//get aspect ratio of video
	aspect_ratio, err := getVideoAspectRatio(tmp.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get aspect ratio", err)
		return
	}
	aspectType := aspectRatioType(aspect_ratio)

	//reset file pointer to read file from beginning again
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to return to beginning of file", err)
		return
	}

	//insert file into s3 bucket
	file_name := MakeFileName()
	file_name = fmt.Sprintf("%s/%s.mp4", aspectType, file_name)

	putInput := &s3.PutObjectInput{
		Bucket:			&cfg.s3Bucket,
		Key:			&file_name,
		Body:			file,
		ContentType:	&mediaType,
	}
	if _, err = cfg.s3Client.PutObject(r.Context(), putInput); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to upload to s3 bucket", err)
		return
	}

	//update VideoURL in database records
	vidURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, file_name)
	updatedVid := database.Video{
		ID:					vid.ID,
		CreatedAt:			vid.CreatedAt,
		UpdatedAt:			vid.UpdatedAt,
		ThumbnailURL:		vid.ThumbnailURL,
		VideoURL:			&vidURL,
		CreateVideoParams:	database.CreateVideoParams{
			Title:			vid.Title,
			Description:	vid.Description,
			UserID:			userID,
		},
	}
	err = cfg.db.UpdateVideo(updatedVid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating database", err)
		return
	}
}