package settings

import (
	"fmt"
	"os"
	"regexp"

	"github.com/766b/mobi"
	"github.com/BurntSushi/toml"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Config struct {
	Title           string
	Cover           string
	Thumbnail       string
	Author          string
	Chapter         string
	SubChapter      string
	Encoding        string
	File            string
	ChapterRegex    *regexp.Regexp
	SubChapterRegex *regexp.Regexp
	Compress        bool
	decode          *encoding.Decoder
}

func New(title, cover, thumbnail, author, chapter, subchapter, encoding, file string, compress bool) (*Config, error) {
	config := &Config{
		Title:           title,
		Cover:           cover,
		Thumbnail:       thumbnail,
		Author:          author,
		Chapter:         chapter,
		SubChapter:      subchapter,
		Encoding:        encoding,
		File:            file,
		Compress:        compress,
		ChapterRegex:    nil,
		SubChapterRegex: nil,
		decode:          nil,
	}
	err := config.Check()
	return config, err
}

func (config *Config) Check() (err error) {
	switch config.Encoding {
	case "GB18030", "gb18030":
		config.decode = simplifiedchinese.GB18030.NewDecoder()
	case "GBK", "gbk":
		config.decode = simplifiedchinese.GBK.NewDecoder()
	case "UTF8", "utf8", "utf-8", "":
		config.decode = nil
	default:
		return fmt.Errorf("Unsupport encoding[GB18030,GBK,UTF8(default)]:%s", config.Encoding)
	}
	if _, err = os.Stat(config.File); os.IsNotExist(err) {
		return
	}
	config.ChapterRegex, err = regexp.Compile(config.Chapter)
	if err == nil && config.SubChapter != "" {
		config.SubChapterRegex, err = regexp.Compile(config.SubChapter)
	}
	return
}

func NewConfig(configFile string) (config *Config, err error) {
	config = &Config{}
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		return
	}
	err = config.Check()
	return
}

func (config *Config) NewWriter(fileName string) (*mobi.MobiWriter, error) {
	if fileName == "" {
		fileName = config.Title + ".mobi"
	}
	m, err := mobi.NewWriter(fileName)
	if err != nil {
		return nil, err
	}
	m.Title(config.Title)
	if !config.Compress {
		m.Compression(mobi.CompressionNone)
	}
	if config.Cover != "" && config.Thumbnail != "" {
		m.AddCover(config.Cover, config.Thumbnail)
	}
	m.NewExthRecord(mobi.EXTH_DOCTYPE, "EBOK")
	m.NewExthRecord(mobi.EXTH_LANGUAGE, "zh")
	m.NewExthRecord(mobi.EXTH_AUTHOR, config.Author)
	return m, nil
}

func (c *Config) Decode(content []byte) ([]byte, error) {
	if c.decode != nil {
		return c.decode.Bytes(content)
	}
	return content, nil
}
