package settings

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"unicode/utf8"

	"golang.org/x/image/font"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"

	"github.com/766b/mobi"
	"github.com/BurntSushi/toml"
	"github.com/flopp/go-findfont"
	"github.com/golang/freetype/truetype"
	z "github.com/nutzam/zgo"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func findFontExpected(fontExpected []string) (fontPath string, err error) {
	for _, fontE := range fontExpected {
		fontPath, err = findfont.Find(fontE)
		if err == nil {
			return
		}
	}
	return fontPath, fmt.Errorf("Font not Found")
}

func LoadFont() (font *truetype.Font, err error) {
	fontExpected := []string{"DroidSansFallbackLegacy.ttf", "SourceHanSansSC-Bold.ttf", "NotoSansMono-Bold.ttf", "NotoSansCJK-Bold.ttc", "simhei.ttf", "simsunb.ttf"}
	fontPath, err := findFontExpected(fontExpected)
	if err != nil {
		return
	}
	fontData, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return
	}
	return truetype.Parse(fontData)
}

type Config struct {
	Title             string
	Cover             string
	Thumbnail         string
	Author            string
	Chapter           string
	SubChapter        string
	Encoding          string
	File              string
	ChapterRegex      *regexp.Regexp
	SubChapterRegex   *regexp.Regexp
	Compress          bool
	decode            *encoding.Decoder
	Font              *truetype.Font
	DefaultBackground image.Image
}

func New(title, cover, thumbnail, author, chapter, subchapter, encoding, file string, compress bool) *Config {
	return &Config{
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
}

func (c *Config) Update(file, title, author, cover, thumbnail string) {
	if file != "" {
		c.File = file
	}
	if title != "" {
		c.Title = title
	}
	if author != "" {
		c.Author = author
	}
	if cover != "" {
		c.Cover = cover
	}
	if thumbnail != "" {
		c.Thumbnail = thumbnail
	}
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
		return fmt.Errorf("text file:%v", err)
	}
	config.ChapterRegex, err = regexp.Compile(config.Chapter)
	if err == nil && config.SubChapter != "" {
		config.SubChapterRegex, err = regexp.Compile(config.SubChapter)
	}
	if err != nil {
		return fmt.Errorf("regexp Compile:%v", err)
	}
	if config.Cover != "" && config.Thumbnail == "" {
		img, err := gg.LoadImage(config.Cover)
		if err != nil {
			return fmt.Errorf("cover:%v", err)
		}
		config.Thumbnail = config.Cover
		width := img.Bounds().Dx()
		if width < 360 {
			config.Cover, err = ScaleImage(config.Cover, 180)
		} else {
			config.Thumbnail, err = ScaleImage(config.Cover, 180)
		}

	}
	return
}

func NewConfig(configFile string) (config *Config, err error) {
	config = &Config{}
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		return
	}
	return
}

func (c *Config) DefaultCover() bool {
	if c.Cover == "" {
		c.Cover = c.Title + "_cover.png"
		c.Thumbnail = c.Title + "_thumbnail.png"
		return true
	}
	return false
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
	m.NewExthRecord(mobi.EXTH_AUTHOR, config.Author)
	return m, nil
}

func (c *Config) Decode(content []byte) ([]byte, error) {
	if c.decode != nil {
		return c.decode.Bytes(content)
	}
	return content, nil
}

func (c *Config) CreateDefaultCover() (err error) {
	rec := c.DefaultBackground.Bounds()
	dc := gg.NewContext(rec.Dx(), rec.Dy())
	dc.DrawImage(c.DefaultBackground, 0, 0)
	dc.Fill()
	err = c.drawTitle(dc, &rec)
	if err != nil {
		return
	}
	err = dc.SavePNG(c.Cover)
	if err != nil {
		return
	}
	return c.createThumbnail()
}

func (c *Config) createThumbnail() (err error) {
	img, err := gg.LoadImage(c.Cover)
	if err != nil {
		return
	}
	out, err := os.Create(c.Thumbnail)
	if err != nil {
		return
	}
	defer out.Close()
	rec := img.Bounds()
	m := resize.Resize(uint(rec.Dx()/5), uint(rec.Dy()/5), img, resize.MitchellNetravali)
	return png.Encode(out, m)
}

func (c *Config) loadFontFace(size float64) font.Face {
	face := truetype.NewFace(c.Font, &truetype.Options{
		Size: size,
		// Hinting: font.HintingFull,
	})
	return face
}

func (c *Config) drawTitle(dc *gg.Context, rec *image.Rectangle) error {
	wight := float64(rec.Dx())
	hight := float64(rec.Dy())
	strLength := float64(utf8.RuneCountInString(c.Title))
	fontSize := math.Min(wight*8/9/strLength, wight/7)
	font := c.loadFontFace(fontSize)
	dc.SetRGB255(47, 79, 79)
	dc.SetFontFace(font)
	w, _ := dc.MeasureString(c.Title)
	dc.DrawString(c.Title, wight/2-w/2, hight/4)
	dc.Stroke()
	if c.Author != "" {
		authorStr := "--by " + c.Author
		strLength = float64(utf8.RuneCountInString(authorStr))
		fontSize = math.Min(wight*2/3/strLength, fontSize*0.8)
		font = c.loadFontFace(fontSize)
		dc.SetFontFace(font)
		dc.SetRGB255(0, 0, 0)
		w, _ = dc.MeasureString(authorStr)
		dc.DrawString(authorStr, wight*8/9-w, hight*5/12)
		dc.Fill()
	}

	return nil
}

func ScaleImage(src string, width int) (fileName string, err error) {
	var img image.Image
	imageType := z.FileType(src)
	switch imageType {
	case "jpeg", "jpg":
		fileName = src + "_scare.jpg"
		img, err = z.ImageJPEG(src)
	case "png", "PNG":
		fileName = src + "_scare.png"
		img, err = z.ImagePNG(src)
	default:
		err = fmt.Errorf("support: .jpg,.png")
	}
	if err != nil {
		return
	}
	bound := img.Bounds()
	dx := bound.Dx()
	dy := bound.Dy()
	m := resize.Resize(uint(width), uint(width*dy/dx), img, resize.MitchellNetravali)
	switch imageType {
	case "jpg", "jpeg":
		err = z.ImageEncodeJPEG(fileName, m, 90)
	case "png", "PNG":
		err = z.ImageEncodePNG(fileName, m)
	}
	return fileName, err
}
