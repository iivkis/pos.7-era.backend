package config

var Flags struct {
	Port *string
	Main *bool
}

var Env struct {
	ServerName string `mapstructure:"SERVER_NAME"`

	OutProtocol string `mapstructure:"OUT_PROTOCOL"`
	OutHost     string `mapstructure:"OUT_HOST"`
	OutPort     string `mapstructure:"OUT_PORT"`

	TokenSecretKey string `mapstructure:"TOKEN_SECRET_KEY"`

	EmailLogin    string `mapstructure:"EMAIL_LOGIN"`
	EmailPassword string `mapstructure:"EMAIL_PASSWORD"`

	DatabaseIP       string `mapstructure:"DATABASE_IP"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
	DatabaseLogin    string `mapstructure:"DATABASE_LOGIN"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`

	SelectelS3AccessKey  string `mapstructure:"SELECTEL_S3_ACCESS_KEY"`
	SelectelS3SecretKey  string `mapstructure:"SELECTEL_S3_SECRET_KEY"`
	SelecletS3BacketName string `mapstructure:"SELECTEL_S3_BACKET_NAME"`
}

var File struct {
	EmailTemplatesDir string
}
