package main

import (
	"flag"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/larspensjo/config"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	deviceFile = flag.String("e", "device.ini", "device file")
	configFile = flag.String("c", "config.ini", "General configuration file")
	uploadDir  = flag.String("w", "./", "Upload work directory")
	isDelete   = flag.Bool("d", false, "When upload finished, delete local file")
)

type IOParamters struct {
	deviceId     string
	ossEndPoint  string
	ossBucket    string
	ossAccessId  string
	ossAccessKey string
}

func (p *IOParamters) parseDevice(iniPath string) {
	cfg, err := config.ReadDefault(iniPath)

	if err != nil {
		log.Fatalf("Fail to find %v,%v", iniPath, err)
	}

	if cfg.HasSection("device") {

		tempValue, err := cfg.String("device", "device_id")
		if err == nil {
			p.deviceId = tempValue
		}
	}
}

func (p *IOParamters) parseOssConfig(iniPath string) {
	cfg, err := config.ReadDefault(iniPath)

	if err != nil {
		log.Fatalf("Fail to find %v,%v", iniPath, err)
	}

	if cfg.HasSection("upload") {

		tempValue2, err := cfg.String("upload", "oss_end_point")
		if err == nil {
			p.ossEndPoint = tempValue2
		}

		tempValue3, err := cfg.String("upload", "oss_bucket")
		if err == err {
			p.ossBucket = tempValue3
		}

		tempValue4, err := cfg.String("upload", "oss_access_id")
		if err == nil {
			p.ossAccessId = tempValue4
		}

		tempValue5, err := cfg.String("upload", "oss_access_key")
		if err == nil {
			p.ossAccessKey = tempValue5
		}
	}
}

func getUploadFiles(uploadDir string) []string {
	dirList, e := ioutil.ReadDir(uploadDir)
	if e != nil {
		fmt.Println("read dir error")
		return nil
	}

	var files []string
	for _, v := range dirList {
		if !v.IsDir() {
			files = append(files, v.Name())
		}
	}

	return files
}

func deleteFile(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println(err)
	}
}

type OssProgressListener struct {
	fileName string
}

func (listener *OssProgressListener) ProgressChanged(event *oss.ProgressEvent) {
	switch event.EventType {
	case oss.TransferStartedEvent:
		fmt.Printf("Transfer %s Started, ConsumedBytes: %d, TotalBytes %d.\n",
			listener.fileName, event.ConsumedBytes, event.TotalBytes)
	case oss.TransferDataEvent:
		fmt.Printf("\rTransfer %s Data, ConsumedBytes: %d, TotalBytes %d, %d%%.", listener.fileName,
			event.ConsumedBytes, event.TotalBytes, event.ConsumedBytes*100/event.TotalBytes)
	case oss.TransferCompletedEvent:
		fmt.Printf("\nTransfer %s Completed, ConsumedBytes: %d, TotalBytes %d.\n",
			listener.fileName, event.ConsumedBytes, event.TotalBytes)
	case oss.TransferFailedEvent:
		fmt.Printf("\nTransfer %s Failed, ConsumedBytes: %d, TotalBytes %d.\n",
			listener.fileName, event.ConsumedBytes, event.TotalBytes)
	default:
	}
}

type OSSUpLoad struct {
	endPoint    string
	bucketName  string
	accessId    string
	accessKey   string
	client      *oss.Client
	bucket      *oss.Bucket
	isConnected bool
}

func (upload *OSSUpLoad) initOssUpload() {
	client, err := oss.New(upload.endPoint, upload.accessId, upload.accessKey)
	if err != nil {
		fmt.Println("Error:", err)
		upload.isConnected = false
		return
	}

	upload.client = client

	bucket, err := upload.client.Bucket(upload.bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		upload.isConnected = false
		return
	}

	upload.bucket = bucket
	upload.isConnected = true
}

func (upload *OSSUpLoad) uploadFile(destFile string, localFile string) bool {
	if upload.isConnected {
		var listener OssProgressListener
		listener.fileName = localFile
		//err := upload.bucket.PutObjectFromFile(destFile, localFile, oss.Progress(&listener))
		err := upload.bucket.UploadFile(destFile, localFile, 100*1024, oss.Progress(&listener), oss.Routines(3), oss.Checkpoint(true, ""))
		return err == nil
	}

	return false
}

func main() {
	flag.Parse()

	var paramters IOParamters
	paramters.parseDevice(*deviceFile)
	paramters.parseOssConfig(*configFile)
	fmt.Println(paramters)

	files := getUploadFiles(*uploadDir)

	var upload OSSUpLoad
	upload.endPoint = paramters.ossEndPoint
	upload.bucketName = paramters.ossBucket
	upload.accessId = paramters.ossAccessId
	upload.accessKey = paramters.ossAccessKey
	upload.initOssUpload()

	for i, file := range files {

		localPath := *uploadDir + "/" + file
		destPath := paramters.deviceId + "/" + file

		retry := 5
		status := upload.uploadFile(destPath, localPath)
		for !status && retry > 0 {
			retry--
			fmt.Println("Sleeep 60s to retry")
			time.Sleep(time.Duration(60) * time.Second)

			upload.initOssUpload()
			status = upload.uploadFile(destPath, localPath)
		}

		fmt.Println(fmt.Sprintf("Upload %d: %-16s status: %v", i, file, status))

		if status && *isDelete {
			fmt.Println(fmt.Sprintf("Upload %s finished, to delete it", localPath))
			deleteFile(localPath)
		}
	}
}
