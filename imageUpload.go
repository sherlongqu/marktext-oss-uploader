package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Config struct {
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.json")

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config.json from %s: %v\n", configPath, err)
		os.Exit(1)
	}

	var cfg Config
	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config.json: %v\n", err)
		os.Exit(1)
	}

	if cfg.AccessKeyId == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" || cfg.Endpoint == "" {
		fmt.Fprintln(os.Stderr, "Error: Incomplete configuration in config.json. All fields are required.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Image path not detected")
		os.Exit(1)
	}
	localImagePath := strings.Join(os.Args[1:], " ")

	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyId, cfg.AccessKeySecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Client creation failed: %v\n", err)
		os.Exit(1)
	}

	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to acquire Bucket: %v\n", err)
		os.Exit(1)
	}

	ext := filepath.Ext(localImagePath)
	ossFileName := fmt.Sprintf("markdown_imgs/%d%s", time.Now().UnixNano(), ext)

	err = bucket.PutObjectFromFile(ossFileName, localImagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Upload failed: %v\n", err)
		os.Exit(1)
	}

	pureEndpoint := cfg.Endpoint
	pureEndpoint = strings.TrimPrefix(pureEndpoint, "https://")
	pureEndpoint = strings.TrimPrefix(pureEndpoint, "http://")
	imageUrl := fmt.Sprintf("https://%s.%s/%s", cfg.BucketName, pureEndpoint, ossFileName)

	fmt.Println(imageUrl)
}
