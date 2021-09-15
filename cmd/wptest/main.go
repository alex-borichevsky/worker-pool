package main

import (
	"encoding/json"
	"flag"
	"fmt"
	ex "github.com/borichevskiy/expression_generator"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	DefaultGoroutines       = 5
	DefaultExpressionLength = 3
	BaseURL                 = "http://192.168.99.100:8080/evaluate/"
)

type resp struct {
	gorm.Model
	Expr   string
	Result int
	Errors string
}

type unmarshalResp struct {
	Expr string `json:"expr"`
	Res  int    `json:"res"`
	Err  string `json:"err"`
}

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := initConfig(); err != nil {
		logger.Fatal("error initializing configs:",
			zap.Error(err))
	}

	start := time.Now()

	logger.Info("connecting")

	db, err := gorm.Open("postgres", pgConfig())
	if err != nil {
		logger.Panic("failed to connect database:",
			zap.Error(err))
	}

	defer db.Close()
	db.AutoMigrate(&resp{})

	g := flag.Int("g", DefaultGoroutines, "number of goroutines ")

	c := flag.Int("c", DefaultExpressionLength, "length of expression")

	flag.Parse()

	var wg sync.WaitGroup

	for i := 0; i < *g; i++ {
		wg.Add(1)

		go calculate(uint(*c), db, &wg)
	}

	wg.Wait()

	elapsed := time.Since(start)

	logger.Info("took ===============>",
		zap.Duration("", elapsed),
	)
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}

func pgConfig() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		viper.GetString("db.host"),
		viper.Get("db.port"),
		viper.GetString("db.username"),
		viper.GetString("db.password"),
		viper.GetString("db.dbname"))
}

func calculate(exprLen uint, db *gorm.DB, wg *sync.WaitGroup) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	defer wg.Done()

	expr := ex.Generate(exprLen)

	logger.Info("requested expression is: ",
		zap.String("expr", expr))

	URL, err := url.Parse(BaseURL)

	if err != nil {
		logger.Error("URL parsing failed:",
			zap.Error(err))
	}

	params := url.Values{}

	params.Add("expr", expr)

	URL.RawQuery = params.Encode()

	logger.Info(URL.String())

	resp, err := http.Get(URL.String())

	if err != nil {
		logger.Error("failed to fetch URL",
			zap.Error(err))
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logger.Error(err.Error())
	}

	defer resp.Body.Close()

	res := parseResp(body)

	db.Create(&res)

	logger.Info(string(body))
}

func parseResp(data []byte) resp {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var ur unmarshalResp
	err := json.Unmarshal(data, &ur)
	if err != nil {
		logger.Fatal("unmarshal json error: ",
			zap.Error(err))
	}

	return resp{
		Expr:   ur.Expr,
		Result: ur.Res,
		Errors: ur.Err,
	}
}
