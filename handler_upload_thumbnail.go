package main

import (
	"fmt"
	"net/http"
	"io"
	"path/filepath"
	"os"
	"crypto/rand"
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

	//get file extension from mediaType
	if mediaType[:6] == "image/" {
		mediaType = mediaType[6:]
	} else {
		respondWithError(w, http.StatusBadRequest, "Media not an image", err)
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

	//randomize file_name
	key := make([]byte, 32)
	rand.Read(key)
	file_name := base64.RawURLEncoding.EncodeToString(key)

	//save img_data at /assets/
	full_file_name := fmt.Sprintf("%s.%s", file_name, mediaType)
	asset_path := filepath.Join(cfg.assetsRoot,"/",full_file_name)
	
	tn, err := os.Create(asset_path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "File creation failed", err)
		return
	}
	if _, err := io.Copy(tn, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "File can't be copied", err)
		return
	}

	//turn img_str into ThumbnailURL
	new_thumbnail_url := fmt.Sprintf("http://localhost:8091/%s", asset_path)
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
