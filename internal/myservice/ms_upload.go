package myservice

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/config"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"github.com/iivkis/pos.7-era.backend/internal/selectelS3Cloud"
)

var UPLOAD_PHOTO_ALLOWED_CONTENT_TYPE = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

type UploadService struct {
	repo     *repository.Repository
	s3cloud  *selectelS3Cloud.SelectelS3Cloud
	alphabet []byte
	rand     *rand.Rand
}

func newUploadService(repo *repository.Repository, s3cloud *selectelS3Cloud.SelectelS3Cloud) *UploadService {
	return &UploadService{
		repo:     repo,
		s3cloud:  s3cloud,
		alphabet: []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"),
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *UploadService) genPhotoName(length int) string {
	name := make([]byte, length)
	for i := range name {
		name[i] = s.alphabet[s.rand.Intn(len(s.alphabet))]
	}
	return string(name)
}

type UploadPhotoInput struct {
	Photo string `form:"photo"`
}

type UploadPhotoOutput struct {
	PhotoID  string `json:"photo_id"`
	PhotoURI string `json:"photo_uri"`
}

//@Summary Загрузить фотографию на сервер
//@param type body UploadPhotoInput false "фото"
//@Accept json
//@Success 201 {object} UploadPhotoOutput "возвращает ссылку на фотографию"
//@Router /upload.Photo [post]
func (s *UploadService) UploadPhoto(c *gin.Context) {
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errUploadFile(err.Error()))
		return
	}
	defer file.Close()

	//get content type
	contentType := header.Header.Get("Content-Type")
	if !UPLOAD_PHOTO_ALLOWED_CONTENT_TYPE[contentType] {
		NewResponse(c, http.StatusBadRequest, errUploadFile("invalid type"))
		return
	}

	claims := mustGetEmployeeClaims(c)
	uploader := s3manager.NewUploader(s.s3cloud.GetSession())

	//generate photo name
	fileID := strconv.Itoa(int(claims.OrganizationID)) + "-" + s.genPhotoName(50)

	//create upload input
	uploadInput := &s3manager.UploadInput{
		Body:        file,
		ACL:         aws.String("public-read"),
		Bucket:      aws.String(config.Env.SelecletS3BacketName),
		Key:         aws.String(fileID),
		ContentType: aws.String(contentType),
	}

	//upload photo
	if _, err = uploader.Upload(uploadInput); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, UploadPhotoOutput{PhotoID: fileID, PhotoURI: s.s3cloud.GetURIFromFileID(fileID)})
}
