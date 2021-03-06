package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var File struct {
	EmailTmplDir string
}

var Env struct {
	ServerName string

	OutProtocol string
	OutHost     string
	OutPort     string
	Secret      string

	EmailLogin string
	EmailPwd   string

	DatabaseIP       string
	DatabaseName     string
	DatabaseLogin    string
	DatabasePassword string

	SelectelS3AccessKey  string
	SelectelS3SecretKey  string
	SelecletS3BacketName string
}

var Flags struct {
	Port *string
	Main *bool
}

func init() {
	loadEnv()
	loadFlags()
	loadJSON("./config.json")
}

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
		File.EmailTmplDir = getField("email_tmpl_dir")
	}
}

func loadEnv() {

	getEnv := func(envName string) string {
		s, ok := os.LookupEnv(envName)
		if !ok {
			panic(fmt.Sprintf("%s undefined", envName))
		}
		return s
	}

	//secret key & server data
	Env.OutProtocol = getEnv("POSN_OUT_PROTOCOL")
	Env.OutHost = getEnv("POSN_OUT_HOST")
	Env.OutPort = getEnv("POSN_OUT_PORT")
	Env.ServerName = getEnv("POSN_SERVER_NAME")

	//secret JWT key
	Env.Secret = getEnv("POSN_SECRET")

	//env for email
	Env.EmailLogin = getEnv("POSN_EMAIL_LOGIN")
	Env.EmailPwd = getEnv("POSN_EMAIL_PWD")

	//database
	Env.DatabaseIP = getEnv("POSN_DATABASE_IP")
	Env.DatabaseName = getEnv("POSN_DATABASE_NAME")
	Env.DatabaseLogin = getEnv("POSN_DATABASE_LOGIN")
	Env.DatabasePassword = getEnv("POSN_DATABASE_PWD")

	//selectel s3 cloud
	Env.SelectelS3AccessKey = getEnv("POSN_SEL_S3_ACCESS_KEY")
	Env.SelectelS3SecretKey = getEnv("POSN_SEL_S3_SECRET_KEY")
	Env.SelecletS3BacketName = getEnv("POSN_SEL_S3_BACKET")
}

func loadFlags() {
	port, _ := os.LookupEnv("PORT")

	Flags.Port = flag.String("port", port, "server port (default from env `PORT`)")
	Flags.Main = flag.Bool("main", false, "main server make db migration, invites cleaning and other functions")

	flag.Parse()
}
