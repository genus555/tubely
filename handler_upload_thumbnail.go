package main

import (
	"fmt"
	"net/http"
	"io"
	"encoding/base64"

	"github.com/genus555/tubely/internal/auth"
	"github.com/genus555/tubely/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

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


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	//parse form data
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	//get media type
	vals, ok := header.Header["Content-Type"]
	if !ok || len(vals) == 0 {
		respondWithError(w, http.StatusNotFound, "Header not found", err)
		return
	}
	mediaType := vals[0]

	//get image data
	img_data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read file", err)
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

	//turn thumbnail data into str
	img_str := base64.StdEncoding.EncodeToString(img_data)

	//turn img_str into ThumbnailURL
	new_thumbnail_url := fmt.Sprintf("data:%s;base64,%s", mediaType, img_str)
	video := database.Video{
		ID:					videoID,
		ThumbnailURL:		&new_thumbnail_url,
		VideoURL:			vid.VideoURL,
		CreateVideoParams:	database.CreateVideoParams{
			Title:				vid.Title,
			Description:		vid.Description,
			UserID:				vid.UserID,
		},
	}
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating video", err)
		return
	}
	respondWithJSON(w, http.StatusOK, video)

	// TODO END
}
