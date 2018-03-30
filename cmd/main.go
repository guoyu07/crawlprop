package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/millken/crawlprop/api"
	"github.com/millken/crawlprop/config"
	"github.com/millken/crawlprop/core"
	"github.com/millken/crawlprop/stored"
)

const version = "1.0.0"

var (
	flagConfigFile = flag.String("c", "./config.json", "Path to config.")
)

func init() {

	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	rand.Seed(time.Now().UnixNano())

	config.Version = version

}

func main() {
	log.Printf("crawlprop v%s // by millken\n", version)
	flag.Parse()

	var cfg config.Config

	data, err := ioutil.ReadFile(*flagConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = config.Decode(string(data), &cfg, "toml"); err != nil {
		log.Fatal(err)
	}
	filterWriter := os.Stdout
	if cfg.Logging.Output != "" {

		switch cfg.Logging.Output {
		case "stdout":
			filterWriter = os.Stdout
		case "stderr":
			filterWriter = os.Stderr

		default:
			filterWriter, err = os.Create(cfg.Logging.Output)
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}

		}

	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(cfg.Logging.Level)),
		Writer:   filterWriter,
	}

	log.SetOutput(filter)
	//log.SetFlags(log.LstdFlags | log.Lshortfile)

	go api.Start(cfg.Api)

	if err = stored.Initialize(cfg.Redis); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
	core.Initialize(cfg)
	<-(chan string)(nil)

}
