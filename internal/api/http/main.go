package apihttp

// import (
// 	"github.com/gin-gonic/gin"
// 	controllerV1 "github.com/iivkis/pos.7-era.backend/internal/api/http/controllers/v1"
// 	"github.com/iivkis/pos.7-era.backend/internal/repository"
// 	"github.com/iivkis/pos.7-era.backend/internal/s3cloud"
// 	"github.com/iivkis/pos.7-era.backend/pkg/authjwt"
// 	"github.com/iivkis/pos.7-era.backend/pkg/mailagent"
// 	"github.com/iivkis/strcode"
// )

// type apihttp struct {
// 	engine     *gin.Engine
// 	repo       *repository.Repository
// 	strcode    *strcode.Strcode
// 	postman    *mailagent.MailAgent
// 	tokenMaker *authjwt.AuthJWT
// 	s3cloud    *s3cloud.SelectelS3Cloud
// }

// func New(
// 	repo *repository.Repository,
// 	strcode *strcode.Strcode,
// 	postman *mailagent.MailAgent,
// 	tokenMaker *authjwt.AuthJWT,
// 	s3cloud *s3cloud.SelectelS3Cloud,
// ) *apihttp {
// 	api := &apihttp{
// 		repo:       repo,
// 		strcode:    strcode,
// 		postman:    postman,
// 		tokenMaker: tokenMaker,
// 		s3cloud:    s3cloud,
// 	}

// 	api.init()
// 	return api
// }

// func (api *apihttp) init() {
// 	api.engine = gin.Default()

// 	controllerV1.AddController(api.engine, api.repo, api.strcode, api.postman, api.tokenMaker, api.s3cloud)
// }

// func (api *apihttp) Engine() *gin.Engine {
// 	return api.engine
// }
