package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var File struct {
	EmailTmplDir string
}

var Env struct {
	ServerName string

	Protocol string
	Host     string
	Port     string
	Secret   string

	EmailLogin     string
	EmailPwd       string
	EmailForNotify string

	DatabaseLogin    string
	DatabasePassword string
	DatabaseIP       string
}

func init() {
	loadEnv()
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
	Env.Protocol = getEnv("POSN_PROTOCOL")
	Env.Host = getEnv("POSN_HOST")
	Env.Port = getEnv("POSN_PORT")
	Env.ServerName = getEnv("POSN_SERVER_NAME")
	Env.Secret = getEnv("POSN_SECRET")

	//env for email
	Env.EmailLogin = getEnv("POSN_EMAIL_LOGIN")
	Env.EmailPwd = getEnv("POSN_EMAIL_PWD")
	Env.EmailForNotify = getEnv("POSN_EMAIL_TEST_RCP")

	//database
	Env.DatabaseLogin = getEnv("POSN_DATABASE_LOGIN")
	Env.DatabasePassword = getEnv("POSN_DATABASE_PASSWORD")
	Env.DatabaseIP = getEnv("POSN_DATABASE_IP")
}
