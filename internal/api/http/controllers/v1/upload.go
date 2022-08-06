package controller

import (
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
)

var uploadPhotoACT = uploadAllowedContentType{
	"image/jpeg": struct{}{},
	"image/png":  struct{}{},
}

type upload struct {
	repo    *repository.Repository
	s3cloud *s3cloud.SelectelS3Cloud
}

func newUpload(repo *repository.Repository, s3cloud *s3cloud.SelectelS3Cloud) *upload {
	return &upload{
		repo:    repo,
		s3cloud: s3cloud,
	}
}

func (s *upload) generateID() string {
	return uuid.New().String() + uuid.New().String()
}

type uploadPhotoResponse struct {
	PhotoID  string `json:"photo_id" mapstructure:"photo_id"`
	PhotoURI string `json:"photo_uri" mapstructure:"photo_uri"`
}

// @Summary Загрузить фотографию на сервер
// @Param photo body string true "изображение"
// @Success 201 {object} UploadPhotoOutput "возвращает id фоторгафии в хранилище, ссылку на фотографию"
// @Router /upload.Photo [post]
func (s *upload) UploadPhoto(c *gin.Context) {
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUploadFile(err.Error()))
		return
	}
	defer file.Close()

	//get content type & validation
	contentType := header.Header.Get("Content-Type")
	if !uploadPhotoACT.Allowed(contentType) {
		NewResponse(c, http.StatusBadRequest, errUploadFile("invalid type"))
		return
	}

	claims := mustGetEmployeeClaims(c)
	uploader := s3manager.NewUploader(s.s3cloud.GetSession())

	//generate photo ID
	photoID := strconv.Itoa(int(claims.OrganizationID)) + "-" + s.generateID()

	//create upload input
	uploadInput := &s3manager.UploadInput{
		Body:        file,
		ACL:         aws.String("public-read"),
		Key:         aws.String(photoID),
		Bucket:      aws.String(config.Env.SelecletS3BacketName),
		ContentType: aws.String(contentType),
	}

	//upload photo
	if _, err = uploader.Upload(uploadInput); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	response := &uploadPhotoResponse{
		PhotoID:  photoID,
		PhotoURI: s.s3cloud.GetURIFromFileID(photoID),
	}

	NewResponse(c, http.StatusCreated, response)
}
