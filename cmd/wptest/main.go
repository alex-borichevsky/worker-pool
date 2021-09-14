package main

import (
	"encoding/json"
	"flag"
	"fmt"
	ex "github.com/borichevskiy/expression_generator"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	Err string `json:"err"`
}

// TODO: long function!
// move gorotine to separate function
func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	start := time.Now()

	logrus.Print("connecting")

	db, err := gorm.Open("postgres", pgConfig())
	if err != nil {
		logrus.Panic(err)
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

	logrus.Print("took ===============> %s\n", elapsed)
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
	defer wg.Done()

	expr := ex.Generate(exprLen)

	logrus.Print("requested expression is: %s\n", expr)

	URL, err := url.Parse(BaseURL)

	if err != nil {
		logrus.Error("URL parsing failed")
	}

	params := url.Values{}

	params.Add("expr", expr)

	URL.RawQuery = params.Encode()

	logrus.Print(URL)

	resp, err := http.Get(URL.String())

	if err != nil {
		logrus.Error(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logrus.Error(err)
	}

	defer resp.Body.Close()

	res := parseResp(body)

	db.Create(&res)

	fmt.Println(string(body))
}

func parseResp(data []byte) resp {
	var ur unmarshalResp
	err := json.Unmarshal(data, &ur)
	if err != nil {
		logrus.Fatal(err)
	}

	return resp{
		Expr:   ur.Expr,
		Result: ur.Res,
		Errors: ur.Err,
	}
}
