package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var url string
var contentLength int64
var multiProcessingNumber int64

func main() {
	inputURL()
	printFileSize()
	inputMultiProcessingNumber()

	startTime := time.Now().Unix()
	channelList := make([]chan string, multiProcessingNumber)
	for i := int64(0); i < multiProcessingNumber; i++ {
		channelList[i] = make(chan string, 1)
	}
	multiDownload(channelList)

	mergeFile(channelList)
	completedTime := time.Now().Unix()
	fmt.Printf("耗时:\t%d s", completedTime-startTime)
}
func mergeFile(channelList []chan string) {
	strArray := strings.Split(url, "/")
	fileName := strArray[len(strArray)-1]
	newFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for i := int64(0); i < multiProcessingNumber; i++ {
		temp := <-channelList[i]
		tempFile, _ := os.Open(temp)
		io.Copy(newFile, tempFile)
		os.Remove(temp)
	}
}

func multiDownload(channelList []chan string) {
	tempLength := contentLength / multiProcessingNumber
	start := int64(0)
	end := int64(-1)

	for i := int64(0); i < multiProcessingNumber; i++ {
		start = end + 1;
		end = end + tempLength;
		if i == multiProcessingNumber-1 {
			end = contentLength;
		}
		fmt.Printf("start:\t%d\tend:\t%d\n", start, end)
		go downloadPiece(channelList[i], start, end)
	}
}

func downloadPiece(queue chan string, start int64, end int64) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	fileName := fmt.Sprintf("temp-file-%d-%d", start, end)
	newFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	queue <- fileName

	fmt.Printf("download completed:\t%d\tend:\t%d\n", start, end)
}

func printFileSize() {
	if contentLength > 1024*1024 {
		fmt.Printf("资源大小：%d MB\n", contentLength/1024/1024)
	} else if contentLength > 1024 {
		fmt.Printf("资源大小：%d KB\n", contentLength/1024)
	} else {
		fmt.Printf("资源大小：%d B\n", contentLength)
	}
}

func inputURL() {
	for {
		fmt.Print("请输入下载资源的URL:")
		fmt.Scanln(&url)
		checkResult := checkURL()
		if checkResult == "ok" {
			return
		}
		fmt.Printf("URL检查结果 : [%s]\n", checkResult)
	}
}

func inputMultiProcessingNumber() {
	for {
		fmt.Print("请输入并发数:")
		fmt.Scanln(&multiProcessingNumber)
		if multiProcessingNumber > 0 {
			fmt.Printf("并发数 : [%d]\n", multiProcessingNumber)
			return
		}
		fmt.Printf("并发数必须是数字，且大于0\n")
	}

}

func checkURL() string {
	if len(url) == 0 {
		return "URL不能为空"
	}

	url = strings.Trim(url, " ")
	url = strings.Trim(url, "　")

	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, strings.NewReader(""))
	if err != nil {
		return "请求失败，请确认URL是否正确"
	}
	resp, err := client.Do(req)
	if err != nil {
		return "请求失败，请确认URL是否正确"
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Sprintf("请求失败，ErrorCode : %d", resp.StatusCode)
	}

	if resp.ContentLength == -1 {
		return "当前资源不止多线程下载"
	}

	contentLength = resp.ContentLength

	return "ok"
}
