package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
)

func generateUrlsFromStdin() []string {
	urls := make([]string, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	return urls
}

func downloader(workerNum int, c <-chan string, res chan<- string, wg *sync.WaitGroup) {
	for url := range c {
		err := downloadFile(url)
		if err != nil {
			res <- fmt.Sprintf("[Worker %v] %v failed: %v\n", workerNum, url, err)
			wg.Done()
			continue
		}
		fmt.Printf("[Worker %v] Downloaded %v\n", workerNum, url)
		wg.Done()
	}
	fmt.Printf("[Worker %v] Done.\n", workerNum)
}

func downloadFile(url string) error {
	_, file := path.Split(url)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	var wg sync.WaitGroup
	c := make(chan string)
	urls := generateUrlsFromStdin()
	res := make(chan string, len(urls))
	workerNum := 10
	wg.Add(len(urls))
	go func(urls []string) {
		for i := 0; i < len(urls); i++ {
			c <- urls[i]
		}
	}(urls)

	for i := 0; i < workerNum; i++ {
		go downloader(i, c, res, &wg)
	}
	wg.Wait()
	close(c)
	close(res)
	for r := range res {
		fmt.Printf("%v", r)
	}
	fmt.Println("Done.")
}
