package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
)

func init() {
	loadFlags()
	loadEnv()
	loadJSON("./config.json")
}

func loadFlags() {
	port, _ := os.LookupEnv("PORT")

	Flags.Port = flag.String("port", port, "server port (default from env `PORT`)")
	Flags.Main = flag.Bool("main", false, "main server make db migration, invites cleaning and other functions")

	flag.Parse()
}

func loadEnv() {
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&Env); err != nil {
		panic(err)
	}
}

// func loadEnv() {
// 	getEnv := func(envName string) string {
// 		s, ok := os.LookupEnv(envName)
// 		if !ok {
// 			panic(fmt.Sprintf("%s undefined", envName))
// 		}
// 		return s
// 	}

// 	//secret key & server data
// 	Env.OutProtocol = getEnv("POSN_OUT_PROTOCOL")
// 	Env.OutHost = getEnv("POSN_OUT_HOST")
// 	Env.OutPort = getEnv("POSN_OUT_PORT")
// 	Env.ServerName = getEnv("POSN_SERVER_NAME")

// 	//secret JWT key
// 	Env.TokenSecretKey = getEnv("POSN_SECRET")

// 	//env for email
// 	Env.EmailLogin = getEnv("POSN_EMAIL_LOGIN")
// 	Env.EmailPassword = getEnv("POSN_EMAIL_PWD")

// 	//database
// 	Env.DatabaseIP = getEnv("POSN_DATABASE_IP")
// 	Env.DatabaseName = getEnv("POSN_DATABASE_NAME")
// 	Env.DatabaseLogin = getEnv("POSN_DATABASE_LOGIN")
// 	Env.DatabasePassword = getEnv("POSN_DATABASE_PWD")

// 	//selectel s3 cloud
// 	Env.SelectelS3AccessKey = getEnv("POSN_SEL_S3_ACCESS_KEY")
// 	Env.SelectelS3SecretKey = getEnv("POSN_SEL_S3_SECRET_KEY")
// 	Env.SelecletS3BacketName = getEnv("POSN_SEL_S3_BACKET")
// }

func loadJSON(configFilePath string) {
	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0o777)
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var data map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}

	getField := func(fieldName string) string {
		d, ok := data[fieldName]
		if !ok {
			panic(fmt.Sprintf("%s undefined", fieldName))
		}
		return d
	}

	//require fields
	{
		File.EmailTemplatesDir = getField("email_tmpl_dir")
	}
}
