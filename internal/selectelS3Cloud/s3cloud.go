package selectelS3Cloud

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/iivkis/pos.7-era.backend/internal/config"
)

type SelectelS3Cloud struct {
	sess   *session.Session
	cndURI string
}

func NewSelectelS3Cloud(accessKeyID string, secretKey string, cdnURI string) *SelectelS3Cloud {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String("https://s3.storage.selcloud.ru"),
		Region:           aws.String("ru-1"),
		S3ForcePathStyle: aws.Bool(true),

		Credentials: credentials.NewStaticCredentials(
			accessKeyID,
			secretKey,
			"", //токен создастся сам после успешной авторизации в облаке
		),
	})

	if err != nil {
		panic(err)
	}
	log.Println("success connection to s3.storage.selcloud.ru")

	return &SelectelS3Cloud{
		sess:   sess,
		cndURI: cdnURI,
	}
}

func (s3 *SelectelS3Cloud) GetSession() *session.Session {
	return s3.sess
}

func (s3 *SelectelS3Cloud) GetURIFromFileID(fileID string) string {
	if s3.cndURI != "" {
		return s3.cndURI + "/" + fileID
	}
	return "https://720408.selcdn.ru/" + config.Env.SelecletS3BacketName + "/" + fileID
}
