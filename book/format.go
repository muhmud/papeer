package book

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	epub "github.com/bmaupin/go-epub"
)

func Filename(name string) string {
	filename := name

	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "")

	return filename
}

func ToMarkdown(c chapter) string {

	// make title
	underline := strings.Repeat("=", len(c.Name()))
	title := fmt.Sprintf("%s\n%s", c.Name(), underline)

	// convert content to markdown
	content, err := md.NewConverter("", true, nil).ConvertString(c.Content())
	if err != nil {
		log.Fatal(err)
	}

	// merge title and content
	content = fmt.Sprintf("%s\n\n%s", title, content)

	for _, sc := range c.SubChapters() {
		// merge subchapters
		content = fmt.Sprintf("%s\n\n\n%s", content, ToMarkdown(sc))
	}

	return content
}

func ToEpub(c chapter, filename string) string {
	if len(filename) == 0 {
		filename = fmt.Sprintf("%s.epub", c.Name())
	}

	// init ebook
	e := epub.NewEpub(c.Name())
	e.SetAuthor(c.Author())

	AppendToEpub(e, c, false)

	err := e.Write(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Ebook saved to \"%s\"\n", filename)

	return filename
}

func AppendToEpub(e *epub.Epub, c chapter, imagesOnly bool) {
	content := ""

	if imagesOnly == false {
		content = c.Content()
	}

	// parse content
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(c.Content()))
	if err != nil {
		log.Fatal(err)
	}

	// download images and replace src in img tags of content
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		imagePath, _ := e.AddImage(src, "")

		if imagesOnly {
			imageTag, _ := goquery.OuterHtml(s)
			content += strings.Replace(imageTag, src, imagePath, 1)
		} else {
			content = strings.Replace(content, src, imagePath, 1)
		}
	})

	html := fmt.Sprintf("<h1>%s</h1>%s", c.Name(), content)
	_, err = e.AddSection(html, c.Name(), "", "")
	if err != nil {
		log.Fatal(err)
	}

	for _, sc := range c.SubChapters() {
		AppendToEpub(e, sc, false)
	}
}

func ToMobi(c chapter, filename string) string {
	if len(filename) == 0 {
		filename = fmt.Sprintf("%s.mobi", c.Name())
	} else {

		// add .mobi extension if not specified
		if strings.HasSuffix(filename, ".mobi") == false {
			filename = fmt.Sprintf("%s.mobi", filename)
		}

	}

	filenameEPUB := strings.ReplaceAll(filename, ".mobi", ".epub")
	ToEpub(c, filenameEPUB)

	exec.Command("kindlegen", filenameEPUB).Run()
	// exec command always return status 1 even if it succeed
	// if err != nil {
	// 	log.Fatal(err)
	// }

	fmt.Printf("Ebook saved to \"%s\"\n", filename)

	err := os.Remove(filenameEPUB)
	if err != nil {
		log.Fatal(err)
	}
	
	return filename
}