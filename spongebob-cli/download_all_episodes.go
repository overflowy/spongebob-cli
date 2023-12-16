package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

func downloadAllEpisodes(maxConcurrent int) error {
	dir, err := mkdirSpongebob()
	if err != nil {
		return err
	}

	episodesUrls, _ := getEpisodes()

	// Asynchronously download all episodes but max {maxConcurrent} episode at a time.
	// Source: https://gist.github.com/AntoineAugusti/80e99edfe205baf7a094?permalink_comment_id=4088548#gistcomment-4088548
	sem := semaphore.NewWeighted(int64(maxConcurrent))
	ctx := context.TODO()
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < len(episodesUrls); i++ {
		wg.Add(1)

		go func(i int) {
			_ = sem.Acquire(ctx, 1)
			defer sem.Release(1)
			defer wg.Done()

			episodeUrl := episodesUrls[i]

			videoSource := extractVideo(episodeUrl)

			newFileName, err := getNewFileNameWithDir(dir, i, videoSource)
			if err != nil {
				fmt.Printf("Error while get new file name with dir: %v\n", err)
				return
			}

			fmt.Printf("downloading\t%d %s\n", i, videoSource)

			if err := downloadFile(newFileName, videoSource); err != nil {
				fmt.Printf("failed\t%d %s: %v\n", i, videoSource, err)
				return
			}

			fmt.Printf("success\t\t%d %s\n", i, videoSource)
		}(i)
	}

	wg.Wait()

	fmt.Printf("total time %v\n", time.Since(start))

	return nil
}

func mkdirSpongebob() (string, error) {
	spongebobDir := filepath.Join(".", "spongebob")
	if err := os.MkdirAll(spongebobDir, os.ModePerm); err != nil {
		return "", err
	}

	return spongebobDir, nil
}

func getNewFileNameWithDir(dir string, index int, videoSource string) (string, error) {
	newFileName := fmt.Sprintf("%d-", index)
	newFileName = filepath.Join(dir, newFileName)

	videoSourceSplitted := strings.Split(videoSource, "/")
	if len(videoSourceSplitted) == 0 {
		return "", errors.New("error split video source")
	}

	videoSourceFileName := videoSourceSplitted[len(videoSourceSplitted)-1]
	newFileName += videoSourceFileName
	return newFileName, nil
}

func downloadFile(filepath string, videoUrl string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(videoUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
