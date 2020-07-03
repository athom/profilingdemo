package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/athom/rtrand"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func getInput() (r string) {
	jsonBytes, err := ioutil.ReadFile("data.json")
	if err != nil {
		panic(err)
	}
	r = string(jsonBytes)
	return
}

func toArray(raw gjson.Result) (r []string) {
	for _, item := range raw.Array() {
		r = append(r, item.String())
	}
	return
}

func pickRandom(input []string) (r string) {
	n := len(input)
	i := rtrand.Intn(n)
	return input[i]
}

func generate(topic string, slowFactor int) {
	//var buf *bytes.Buffer
	//buf = bytes.NewBuffer(make([]byte, 0, 1024*1024*200))

	r, _ := generateV2(topic)
	uuid := uuid.New().String()
	filename := fmt.Sprintf("tmp/tmp_file_%v.txt", uuid)
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//r += buf.String()
	for i := 0; i < len(r); i++ {
		s := r[i : i+1]
		_, err := f.Write([]byte(s))
		if err != nil {
			panic(err)
		}

		n := 10
		if i%n == 0 {
			err = f.Sync()
			if err != nil {
				panic(err)
			}
		}

	}

	//_, err = f.WriteString(r)
	//if err != nil {
	//	panic(err)
	//}

	return
}

func generateV2(topic string) (r string, count int) {
	input := getInput()
	before := toArray(gjson.Get(input, "before"))
	after := toArray(gjson.Get(input, "after"))
	famous := toArray(gjson.Get(input, "famous"))
	bosh := toArray(gjson.Get(input, "bosh"))

	for len(r) < 6000 {
		i := rtrand.Intn(100)
		if i < 5 {
			r += "\n    "
			continue
		}
		if i < 20 {
			b := pickRandom(before)
			a := pickRandom(after)
			line := pickRandom(famous)
			line = strings.ReplaceAll(line, "b", b)
			line = strings.ReplaceAll(line, "a", a)
			r += line
			continue
		}
		r += pickRandom(bosh)
	}
	r = strings.ReplaceAll(r, "x", topic)
	count = strings.Count(r, topic)
	return
}

func ok(c *gin.Context) {
	c.JSON(200, "ok")
	return
}

func wordsGenV1(c *gin.Context) {
	trainingCount := 1
	var (
		finalArticle string
		finalCount   int
		err          error
	)
	word := c.Param("word")
	sc := c.Query("samples_count")
	if strings.TrimSpace(sc) != "" {
		trainingCount, err = strconv.Atoi(sc)
		if err != nil {
			panic(err)
		}
	}
	slowFactor := 0
	sf := c.Query("slow_factor")
	if strings.TrimSpace(sf) != "" {
		slowFactor, err = strconv.Atoi(sf)
		if err != nil {
			panic(err)
		}
	}

	_, err = os.Stat("tmp")
	if os.IsNotExist(err) {
		err := os.Mkdir("tmp", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < trainingCount; i++ {
		generate(word, slowFactor)
	}

	var files []string
	err = filepath.Walk("tmp", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "txt") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		bs, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		fullText := string(bs)
		count := strings.Count(fullText, word)
		if finalCount < count {
			finalCount = count
			finalArticle = fullText
		}
	}

	err = os.RemoveAll("tmp")
	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{
		"article":     finalArticle,
		"article_len": len(finalArticle),
		"topic":       word,
		"count":       finalCount,
	})
	return
}

func wordsGenV2(c *gin.Context) {
	trainingCount := 1
	var (
		finalArticle string
		finalCount   int
		err          error
	)
	word := c.Param("word")
	sc := c.Query("samples_count")
	if sc != "" {
		trainingCount, err = strconv.Atoi(sc)
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < trainingCount; i++ {
		article, count := generateV2(word)
		if finalCount < count {
			finalCount = count
			finalArticle = article
		}
	}
	c.JSON(200, gin.H{
		"article":     finalArticle,
		"article_len": len(finalArticle),
		"topic":       word,
		"count":       finalCount,
	})
	return
}

func main() {
	router := gin.Default()
	router.GET("/genv1/:word", wordsGenV1)
	router.GET("/gen/:word", wordsGenV2)
	router.GET("/", ok)
	router.Run(":12000")
}
