package api

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"spMDOImages/server/conf"
	"strconv"
	"time"
)

func writeFiles(entries *[]FileEntry) error {
	// Создаем временную папку и пишем файлы в нее
	imgFolder, err := filepath.Abs(conf.ImagesFolder)
	if err != nil {
		return fmt.Errorf("couldn't get absolute path of images root %s -> %s", imgFolder, err)
	}
	tempFolder := filepath.Join(imgFolder, "/temp")
	if err := os.MkdirAll(tempFolder, 0755); err != nil {
		return fmt.Errorf("couldn't create directory %s -> %s", tempFolder, err)
	}

	if err := writeToTemp(entries, tempFolder); err != nil {
		return fmt.Errorf("couldn't write to temp folder %s -> %s", tempFolder, err)
	}

	if err := handleFoldersNum(imgFolder); err != nil {
		return fmt.Errorf("couldn't handle folders %s", err)
	}
	// Делаем из временной папки постоянную
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	newDirPath := filepath.Join(imgFolder, timestamp)
	if err = os.Rename(tempFolder, newDirPath); err != nil {
		return fmt.Errorf("couldn't rename temp dir to persistent %s", err)
	}
	return nil
}

func getFilePaths() ([]string, error) {
	paths := make([]string, 0)
	picsDir, err := filepath.Abs(conf.ImagesFolder)
	if err != nil {
		return nil, fmt.Errorf("couldn't get absolute path of %s -> %s", conf.ImagesFolder, err)
	}
	folders, err := ioutil.ReadDir(picsDir)
	if err != nil {
		return nil, fmt.Errorf("couldn't read pictures directory %s -> %s", picsDir, err)
	}
	if len(folders) == 0 {
		return []string{}, nil
	}

	lastFolderName := folders[len(folders)-1].Name()
	var lastFolder string
	switch lastFolderName {
	case "temp":
		lastFolder = filepath.Join(picsDir, folders[len(folders)-2].Name())
	default:
		lastFolder = filepath.Join(picsDir, folders[len(folders)-1].Name())
	}

	pics, err := ioutil.ReadDir(lastFolder)
	if err != nil {
		log.Printf("Error while reading pictures directory %s", err)
		return nil, fmt.Errorf("couldn't read pictures directory %s -> %s", lastFolder, err)
	}
	for _, p := range pics {
		picturePath := filepath.Join(lastFolderName, p.Name())
		paths = append(paths, picturePath)
	}
	return paths, nil
}

func makeHash(paths []string) (string, error) {
	var allFiles []byte
	imgFolder, err := filepath.Abs(conf.ImagesFolder)
	if err != nil {
		return "", fmt.Errorf("couldn't get absolute path of images root %s -> %s", imgFolder, err)
	}
	for _, p := range paths {
		picPath := filepath.Join(imgFolder, p)
		if err != nil {
			return "", fmt.Errorf("couldn't get absolute path of picture %s", err)
		}
		content, err := os.Open(picPath)
		if err != nil {
			return "", fmt.Errorf("couldn't open file %s for hashing %s", picPath, err)
		}
		fileBytes, err := ioutil.ReadAll(content)
		if err != nil {
			log.Printf("can't read file %s for hashing", picPath)
			return "", fmt.Errorf("couldn't read file %s for hashing %s", picPath, err)
		}
		allFiles = append(allFiles, fileBytes...)
	}
	byteHash := md5.Sum(allFiles)
	hash := hex.EncodeToString(byteHash[:])
	return hash, nil
}

func writeToTemp(entries *[]FileEntry, tempFolder string) error {
	for _, entry := range *entries {
		path := filepath.Join(tempFolder, entry.Filename)
		content, err := base64.StdEncoding.DecodeString(entry.Content64)
		if err != nil {
			return fmt.Errorf("error while encoding to base64 in SetAdv %s", err)
		}
		if err = ioutil.WriteFile(path, content, 0744); err != nil {
			return fmt.Errorf("couldn't write file %s -> %s", path, err)
		}
	}
	return nil
}

// handleFoldersNum проверяет, сколько папок создано
// и удаляет старые, если их количество превышает максимальное
func handleFoldersNum(path string) error {
	folderContent, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("can't read images directory %s", err)
	}
	if len(folderContent) >= conf.MaxFolderNum+1 {
		foldPath := filepath.Join(path, folderContent[0].Name())
		os.RemoveAll(foldPath)
	}
	return nil
}
