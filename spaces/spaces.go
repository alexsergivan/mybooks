package spaces

import (
	"sync"

	"github.com/alexsergivan/mybooks/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var spacesClient *s3.S3

var once sync.Once

func GetSpacesClient() *s3.S3 {
	once.Do(func() {
		key := config.Config("SPACES_KEY")
		secret := config.Config("SPACES_SECRET")

		s3Config := &aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, ""),
			Endpoint:    aws.String("https://fra1.digitaloceanspaces.com"),
			Region:      aws.String("us-east-1"),
		}

		newSession := session.New(s3Config)
		spacesClient = s3.New(newSession)
	})

	return spacesClient

}
