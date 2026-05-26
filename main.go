package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	mutex sync.Mutex
	// Increase the size of the channel buffer to avoid blocking the goroutine that tracks changes
	processFileChannel = make(chan string, 10)
)

// calculateHash - Calculate the MD5 hash of the file
func calculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// removeDuplicates - Remove duplicate files in the directory
func removeDuplicates(directoryPath string) error {
	hashMap := make(map[string]string)

	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			hash, err := calculateHash(path)
			if err != nil {
				return err
			}

			if existingPath, ok := hashMap[hash]; ok {
				fmt.Printf("Removing duplicate: %s (duplicate of %s)\n", path, existingPath)
				err := os.Remove(path)
				if err != nil {
					return err
				}
			} else {
				hashMap[hash] = path
			}
		}
		return nil
	})

	return err
}

// sortAndRemoveDuplicates - Sort and remove duplicate lines in the file
func sortAndRemoveDuplicates(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	uniqueLines := make(map[string]struct{})
	// use Reader instead previous Scanner, fix: "token too long" error
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		trimmedLine := strings.TrimSpace(line)

		// skip this
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "//") {
			uniqueLines[trimmedLine] = struct{}{}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			file.Close()
			return err
		}
	}
	file.Close()

	// sort
	linesSlice := make([]string, 0, len(uniqueLines))
	for line := range uniqueLines {
		linesSlice = append(linesSlice, line)
	}
	sort.Strings(linesSlice)

	// rewrite
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range linesSlice {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// processFiles - Process all files in the target directory
func processFiles(directoryPath string) error {
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Sort and remove duplicates
			err := sortAndRemoveDuplicates(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// processFile - Process a single file
func processFile(filePath string) {

	fmt.Printf("Processing file: %s\n", filePath)
	err := sortAndRemoveDuplicates(filePath)
	if err != nil {
		fmt.Printf("Error processing file %s: %v\n", filePath, err)
	}
}

// watchDirectory - Track changes in the directory (TESTING)
func watchDirectory(directoryPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//fmt.Printf("Event: %+v\n", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					time.Sleep(1 * time.Second) // Wait for some time to make sure the file is completely written

					mutex.Lock()
					defer mutex.Unlock()
					if _, err := os.Stat(event.Name); err == nil {
						processFile(event.Name)
						//mutex.Unlock()

					}

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Error: %v\n", err)
			}
		}
	}()

	err = filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Initial processing of existing files
			processFile(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	err = watcher.Add(directoryPath)
	if err != nil {
		fmt.Printf("Error adding directory to watcher: %v\n", err)
		return
	}

	fmt.Printf("Watching directory: %s\n", directoryPath)

	select {}
}

func main() {
	var directoryPath string

	watch := false
	version := "0.3.7"

	versionFlag := flag.Bool("version", false, "Print the version of the program")
	pathFlag := flag.String("path", "", "The path to the directory to process")
	watchFlag := flag.Bool("watch", false, "Watch the directory for changes")
	removeFlag := flag.Bool("remove", false, "Remove duplicate files by content hash")

	flag.Parse()

	if *versionFlag {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	if *pathFlag != "" {
		directoryPath = *pathFlag
	} else if flag.NArg() > 0 {
		directoryPath = flag.Arg(0)
	} else {
		fmt.Println("Usage: kompressor <directory-path> [-watch] [-remove]")
		os.Exit(1)
	}

	if *watchFlag {
		watch = true
	}

	if *removeFlag {
		fmt.Println("Duplicate file removal is enabled")
		if err := removeDuplicates(directoryPath); err != nil {
			fmt.Printf("Error removing duplicate files: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Duplicate file removal is disabled")
	}

	if err := processFiles(directoryPath); err != nil {
		fmt.Printf("Error processing files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Sorting and removing duplicate lines completed successfully.")

	if watch {
		go watchDirectory(directoryPath)
		select {}
	}
}
