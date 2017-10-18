package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func writeToTemp(entries *[]FileEntry, path string) error {
	if err := os.Mkdir(path, 0744); err != nil {
		return fmt.Errorf("Error while handling directories %s", err)
	}
	for _, entry := range *entries {
		path := filepath.Join(path, entry.Filename)
		content, err := base64.StdEncoding.DecodeString(entry.Content64)
		if err != nil {
			return fmt.Errorf("error while encoding to base64 in SetAdv %s", err)
		}
		ioutil.WriteFile(path, content, 0744)
	}
	return nil
}

func handleTemp(entries *[]FileEntry) error {
	folderContent, err := ioutil.ReadDir(Config.ImagesRoot)
	if err != nil {
		return fmt.Errorf("can't read images directory %s", err)
	}
	if len(folderContent) >= Config.MaxFolderNum+1 {
		fmt.Println(len(folderContent))
		foldPath := filepath.Join(Config.ImagesRoot, folderContent[0].Name())
		os.RemoveAll(foldPath)
	}
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.00")
	newDirPath := filepath.Join(Config.ImagesRoot, timestamp)
	tempDirPath := filepath.Join(Config.ImagesRoot, "/temp/")
	if err = os.Rename(tempDirPath, newDirPath); err != nil {
		return fmt.Errorf("can't read images directory %s", err)
	}
	return nil
}
