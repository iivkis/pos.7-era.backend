package config

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	onceFlags sync.Once
	onceFiles sync.Once
)

func Load(rootpath string) {
	onceFlags.Do(func() {
		loadFlags()
	})

	onceFiles.Do(func() {
		loadEnv(rootpath)
		loadJSON(rootpath)
	})
}

func loadFlags() {
	port, _ := os.LookupEnv("PORT")

	Flags.Port = flag.String("port", port, "server port (default from env `PORT`)")
	Flags.Main = flag.Bool("main", false, "make db migration, invites cleaning.")

	flag.Parse()
}

func loadEnv(rootpath string) {
	vp := viper.New()

	vp.SetConfigType("env")

	vp.AddConfigPath(rootpath)
	vp.SetConfigName(".env")

	vp.AutomaticEnv()

	for _, env := range os.Environ() {
		split := strings.Split(env, "=")
		vp.SetDefault(split[0], split[1])
	}

	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("`.env` file not found. Hmm.. Am I on Heroku now?")
		} else {
			panic(err)
		}
	}

	if err := vp.Unmarshal(&Env); err != nil {
		panic(err)
	}
}

func loadJSON(rootpath string) {
	vp := viper.New()

	vp.SetConfigType("json")

	vp.AddConfigPath(rootpath)
	vp.SetConfigName("config.json")

	if err := vp.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := vp.Unmarshal(&File); err != nil {
		panic(err)
	}
}
