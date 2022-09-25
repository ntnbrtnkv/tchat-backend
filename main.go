package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const FILES_PATH = "./cache"

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	r := gin.Default()
	nrLicense := os.Getenv("NEWRELIC")

	if nrLicense != "" {
		app, err := newrelic.NewApplication(
			newrelic.ConfigAppName("tchat-backend"),
			newrelic.ConfigLicense(os.Getenv("NEWRELIC")),
			newrelic.ConfigAppLogForwardingEnabled(true),
		)

		if err != nil {
			log.Fatal(err)
		}

		r.Use(nrgin.Middleware(app))
	}

	err := os.MkdirAll(FILES_PATH, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	imageMagicPath := os.Getenv("IMAGE_MAGIC")
	r.Use(cors.Default())
	r.Static("/files", FILES_PATH)
	r.POST("/convert", func(c *gin.Context) {
		url := c.PostForm("url")
		log.Println("url:", url)
		hmd5 := md5.Sum([]byte(url))
		filename := hex.EncodeToString(hmd5[:]) + ".gif"

		filepath := FILES_PATH + "/" + filename

		path := ""

		if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
			// path/to/whatever does not exist
			log.Println("File " + filepath + " not found")

			err := downloadFile(url, filepath)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("File downloaded")

			cmd := exec.Command(imageMagicPath, "-dispose", "none", "-layers", "optimize", filepath, filepath)

			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}

			log.Println("Image converted")
		}
		path = "/files/" + filename

		scheme := "http://"
		if c.Request.TLS != nil {
			scheme = "https://"
		}

		c.JSON(http.StatusOK, gin.H{
			"gif": scheme + c.Request.Host + path,
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		out, err := exec.Command(imageMagicPath, "-version").CombinedOutput()

		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"imagemagick": string(out),
		})
	})
	r.Run()
}
