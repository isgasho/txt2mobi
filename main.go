package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"

	"./settings"
)

var (
	LineFeed       = []byte("<br/>")
	PStart         = []byte("<p>")
	PEnd           = []byte("</p>")
	Blank          = []byte{}
	ConfigFile     = flag.String("f", "", "ebook config file(.toml)")
	IsParagraph    = flag.Bool("p", false, "[option]is to use <p></p>,use false as default")
	OutputFileName = flag.String("o", "", "[option]output file name")
)

type ChapterContent struct {
	Title   string
	Content []byte
}

func (c *ChapterContent) Append(content []byte) {
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
	if *ConfigFile == "" {
		flag.Usage()
		return
	}
	config, err := settings.NewConfig(*ConfigFile)
	if err != nil {
		log.Fatal(err)
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
