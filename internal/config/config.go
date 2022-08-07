package config

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var once sync.Once

func Load(rootpath string) {
	once.Do(func() {
		loadFlags()
		loadEnv(rootpath)
		// loadJSON("./config.json")
	})
}

func loadFlags() {
	port, _ := os.LookupEnv("PORT")

	Flags.Port = flag.String("port", port, "server port (default from env `PORT`)")
	Flags.Main = flag.Bool("main", false, "main server make db migration, invites cleaning and other functions")

	flag.Parse()
}

func loadEnv(rootpath string) {
	viper.SetConfigType("env")
	viper.AddConfigPath(rootpath)
	viper.AutomaticEnv()

	for _, env := range os.Environ() {
		split := strings.Split(env, "=")
		viper.SetDefault(split[0], split[1])
	}

	if _, err := os.Stat(filepath.Join(rootpath, ".env")); os.IsNotExist(err) {
		log.Print(".env file not found")
	} else if err != nil {
		panic(err)
	} else {
		viper.SetConfigName(".env")

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}

	if err := viper.Unmarshal(&Env); err != nil {
		panic(err)
	}
}

// func loadJSON(configFilePath string) {
// 	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0o777)
// 	if err != nil {
// 		panic(err)
// 	}

// 	b, err := ioutil.ReadAll(f)
// 	if err != nil {
// 		panic(err)
// 	}

// 	var data map[string]string
// 	if err := json.Unmarshal(b, &data); err != nil {
// 		panic(err)
// 	}

// 	getField := func(fieldName string) string {
// 		d, ok := data[fieldName]
// 		if !ok {
// 			panic(fmt.Sprintf("%s undefined", fieldName))
// 		}
// 		return d
// 	}

// 	//require fields
// 	{
// 		File.EmailTemplatesDir = getField("email_tmpl_dir")
// 	}
// }
