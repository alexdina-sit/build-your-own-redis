package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func (server *Server) LoadAOF() {
	// Debug - verific daca exista primul DIR si contine ceva
	info, err := os.Stat(server.Config.Dir)
	if info != nil && info.IsDir() {
		fmt.Println("Primul folder", server.Config.Dir, "exista.")
	}

	entries, _ := os.ReadDir(server.Config.Dir)
	fmt.Println("Fisiere existente in primul folder", entries)
	// Sfarsit

	directoryPath := filepath.Join(server.Config.Dir, server.Config.AppendDirname)

	// Debug - verific existenta appendir
	fmt.Println("Path folder appenddir", directoryPath)
	info, err = os.Stat(directoryPath)

	if err != nil {
		fmt.Println("Eroare gasita cand am verificat al doilea folder", err)
	} else {
		fmt.Println("Exista", directoryPath)
	}
	// Sfarsit

	manifestPath := filepath.Join(directoryPath, server.Config.AppendFilename+".manifest")
	aofPath := filepath.Join(directoryPath, server.Config.AppendFilename+".1.incr.aof")
	manifestFile, err := os.Open(manifestPath)

	if err != nil {
		fmt.Println(err)
		if os.IsNotExist(err) {
			err := os.MkdirAll(directoryPath, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}

			aof, _ := createFile(aofPath)
			aof.Close()

			manifestFile, _ = createFile(manifestPath)
			manifestText := fmt.Sprintf("file %s seq 1 type i\n", server.Config.AppendFilename+".1.incr.aof")
			manifestFile.Write([]byte(manifestText))
			readManifestFile(manifestFile)
			manifestFile.Close()

			//Debug - verific daca are continu directorul
			entries, _ = os.ReadDir(directoryPath)
			fmt.Println("Fisiere dupa ce am creat", entries)
			return

		}
	}

	fmt.Println("exista fisierele")
	readManifestFile(manifestFile)
	manifestFile.Close()
}

func createFile(filePath string) (*os.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err

	}

	return file, nil
}

func readManifestFile(file *os.File) (string, int, string, error) {
	fmt.Println("sunt in readmanifest")
	reader := bufio.NewReader(file)
	data, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return "", -1, "", errors.New("Failed to read the manifest file" + err.Error())
		}
		return "", -1, "", nil
	}

	fmt.Println("Continut fisier Manifest:", data)

	re := regexp.MustCompile(`file\s+(.+?)\s+seq\s+(\d+).*type\s+(\w)`)
	match := re.FindStringSubmatch(data)
	if len(match) < 4 {
		return "", -1, "", errors.New("Invalid file content")
	}

	aofFileName := match[1]
	fileType := match[3]
	seq, err := strconv.Atoi(match[2])
	if err != nil {
		return "", -1, "", errors.New("Failed to convert the sequence from string to int")
	}

	fmt.Println(aofFileName, seq, fileType)
	return aofFileName, seq, fileType, nil
}
