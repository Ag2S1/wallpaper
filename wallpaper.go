package main

import (
	"fmt"
	"github.com/antonholmquist/jason"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	_ "strconv"
	"sync"
)

const reddit_url = "http://www.reddit.com/r/earthporn.json"
const folder_name = "wallpapers"
const downloader_count = 4
const min_width = 1080
const min_height = 1920

type Task struct {
	url      string
	filename string
}

func main() {
	dir, err := os.Getwd()
	checkError(err)
	path := dir + "/" + folder_name

	err = os.MkdirAll(path, os.ModePerm)
	checkError(err)
	os.Chdir(path)

	download()
}

func download() {
	resp, err := http.Get(reddit_url)
	checkError(err)
	defer resp.Body.Close()

	data, err := jason.NewObjectFromReader(resp.Body)
	checkError(err)

	images, err := data.GetObjectArray("data", "children")
	checkError(err)

	task_tunnel := make(chan Task)
	var wait_group sync.WaitGroup

	for i := 0; i < downloader_count; i++ {
		go imageDownloader(task_tunnel, &wait_group)
	}

	for _, image := range images {
		url, err := image.GetString("data", "url")
		if err != nil {
			continue
		}
		title, err := image.GetString("data", "title")
		if err != nil {
			continue
		}
		wait_group.Add(1)
		task_tunnel <- Task{url, title + ".jpg"}
	}
	wait_group.Wait()
}

func imageDownloader(task_tunnel chan Task, wait_group *sync.WaitGroup) {
	fmt.Println("New Downloader")
	for {
		task := <-task_tunnel
		downloadImage(task.filename, task.url)
		wait_group.Done()
	}
}

func downloadImage(filename string, url string) {
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("Already exist: " + url)
	}

	fmt.Println("start download: " + url)
	var err error

	defer func() {
		if err == nil {
			fmt.Println("	success " + url)
		} else {
			fmt.Println("	" + err.Error())
		}
	}()

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width >= min_width && height >= min_height {
		out, err := os.Create(filename)
		if err != nil {
			return
		}
		defer out.Close()

		err = jpeg.Encode(out, img, nil)
	}
}

func checkError(err error) {
	if err == nil {
		return
	}
	fmt.Println(err)
	os.Exit(-1)
}
