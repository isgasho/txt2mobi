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
	Title        string
	Cover        string
	Thumbnail    string
	Author       string
	Chapter      string
	SubChapter   string
	Encoding     string
	File         string
	ChapterRegex *regexp.Regexp
	SubChapRegex *regexp.Regexp
	Compress     bool
	decode       *encoding.Decoder
}

func NewConfig(configFile string) (config Config, err error) {
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		return
	}
	switch config.Encoding {
	case "GB18030", "gb18030":
		config.decode = simplifiedchinese.GB18030.NewDecoder()
	case "GBK", "gbk":
		config.decode = simplifiedchinese.GBK.NewDecoder()
	case "UTF8", "utf8", "utf-8", "":
		config.decode = nil
	default:
		return config, fmt.Errorf("Unsupport encoding[GB18030,GBK,UTF8(default)]:%s", config.Encoding)
	}
	if _, err = os.Stat(config.File); os.IsNotExist(err) {
		return
	}

	config.ChapterRegex, err = regexp.Compile(config.Chapter)
	if err != nil {
		return
	}
	if config.SubChapter != "" {
		config.SubChapRegex, err = regexp.Compile(config.SubChapter)
		if err != nil {
			return
		}
	}
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
	m.AddCover(config.Cover, config.Thumbnail)
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
