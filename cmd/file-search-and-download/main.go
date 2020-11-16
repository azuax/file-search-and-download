package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// path where we are going to save the files
const downloadFolder = "downloads"

func filesToDownload(site string) (files []string) {
	// search for links and pass to channel
	doc, err := goquery.NewDocument(site)
	if err != nil {
		log.Println("Can't open site", site)
	}
	s := "html body div#b_content ol#b_results li.b_algo h2 a"
	sel := doc.Find(s)
	for i := range sel.Nodes {
		// external node
		ne := sel.Eq(i)
		for _, ni := range ne.Nodes {
			for _, attr := range (*ni).Attr {
				if attr.Key == "href" {
					files = append(files, attr.Val)
				}
			}
		}
	}
	return files
}

func downloadFile(URL string, wg *sync.WaitGroup, c chan string) {
	defer wg.Done()
	resp, err := http.Get(URL)
	if err != nil {
		log.Println("Can't download", URL)
	}
	defer resp.Body.Close()
	fName := path.Base(resp.Request.URL.String())
	out, err := os.Create(fmt.Sprintf("%s/%s", downloadFolder, fName))
	if err != nil {
		log.Println("Can't download file", fName)
		return
	}
	defer out.Close()

	fmt.Println("Downloading", fName)
	bytes, err := io.Copy(out, resp.Body)
	fmt.Printf("Copied %.2f KB in file %s\n", float32(bytes/(1024)), fName)
	if err != nil {
		log.Println("Can't save the file", fName)
		return
	}
	c <- fName
	return
}

func main() {
	sitePtr := flag.String("s", "", "Site to search for")
	fTypePtr := flag.String("f", "", "Filetype to search for. Example: xslx, docx")
	flag.Parse()
	// first we create the folder if it doesn't exists
	if _, err := os.Stat(downloadFolder); os.IsNotExist(err) {
		os.Mkdir(downloadFolder, 0755)
	}

	wg := new(sync.WaitGroup)

	q := fmt.Sprintf(
		"site:%s && filetype:%s && instreamset:(url title):%s",
		*sitePtr,
		*fTypePtr,
		*fTypePtr,
	)
	bingURL := fmt.Sprintf("http://www.bing.com/search?q=%s", url.QueryEscape(q))

	fToD := filesToDownload(bingURL)

	fmt.Printf("Files to download: %d\n", len(fToD))
	cResult := make(chan string, len(fToD))
	for _, l := range fToD {
		wg.Add(1)
		go downloadFile(l, wg, cResult)
	}
	wg.Wait()
	close(cResult)
	fmt.Println("List of downloaded files: ")
	for fName := range cResult {
		fmt.Printf("\t-%s\n", fName)
	}
}
