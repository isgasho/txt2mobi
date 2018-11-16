package main

import (
	"bufio"
	"bytes"
	"flag"
	"html"
	"log"
	"os"

	"./settings"
)

var (
	LineFeed       = []byte("<br/>")
	PStart         = []byte("<p>")
	PEnd           = []byte("</p>")
	Blank          = []byte{}
	ConfigFile     = flag.String("config", "", "ebook config file(.toml)")
	IsParagraph    = flag.Bool("p", false, "[option]is to use <p></p>,use false as default")
	OutputFileName = flag.String("o", "", "[option]output file name")
	IsEscape       = flag.Bool("escape", false, "[option]To Disable html escape")
	MetaFile       = flag.String("f", "", "input file")
	MetaCover      = flag.String("cover", "", "mobi cover")
	MetaTitle      = flag.String("title", "", "mobi title")
	MetaAuthor     = flag.String("author", "", "EBOK author")
	MetaCompress   = flag.Bool("compress", false, "Is to compress")
	MetaEncoding   = flag.String("encoding", "gb18030", "encoding:gb18030(default),gbk,uft-8")
	MetaChapter    = flag.String("chapter", "^第[零一二三四五六七八九十百千两\\d]+章 .*$", "regexp pattern for chapter,default:'^第[零一二三四五六七八九十百千两\\d]+章 .*$'")
)

type ChapterContent struct {
	Title   string
	Content []byte
}

func (c *ChapterContent) Append(content []byte) {
	if !*IsEscape {
		content = []byte(html.EscapeString(string(content)))
	}
	if *IsParagraph {
		if len(content) > 1 {
			c.Content = append(c.Content, bytes.Join([][]byte{PStart, content, PEnd}, Blank)...)
		} else {
			c.Content = append(c.Content, LineFeed...)
		}
	} else {
		c.Content = append(c.Content, LineFeed...)
		if len(content) > 1 {
			c.Content = append(c.Content, content...)
		}
	}

}

func (c *ChapterContent) SetTitle(title string) {
	c.Title = title
}

func (c *ChapterContent) Restore(title string) {
	c.Title = title
	c.Content = make([]byte, 0)
}

func main() {
	flag.Parse()
	var config *settings.Config
	var err error
	if *ConfigFile != "" {
		config, err = settings.NewConfig(*ConfigFile)
	} else {
		config, err = settings.New(*MetaTitle, *MetaCover, *MetaCover, *MetaAuthor, *MetaChapter, *MetaEncoding, *MetaFile, *MetaCompress)
	}
	if err != nil {
		log.Fatal(err)
		flag.Usage()
		return
	}
	file, err := os.Open(config.File)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	mobiWriter, err := config.NewWriter(*OutputFileName)
	if err != nil {
		log.Fatal(err)
	}

	chapter := ChapterContent{
		Title:   config.Title,
		Content: []byte{},
	}
	regex := config.ChapterRegex
	var line []byte
	for scanner.Scan() {
		line, err = config.Decode(scanner.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		if regex.Match(line) {
			if len(chapter.Content) > 0 {
				mobiWriter.NewChapter(chapter.Title, chapter.Content)
			}
			chapter.Restore(string(line))
		} else {
			chapter.Append(line)
		}
	}
	if len(chapter.Content) > 0 {
		mobiWriter.NewChapter(chapter.Title, chapter.Content)
	}
	mobiWriter.Write()
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
