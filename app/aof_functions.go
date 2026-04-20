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

func (server *Server) LoadAOF() error {
	directoryPath := filepath.Join(server.Config.Dir, server.Config.AppendDirname)

	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		return err
	}

	manifestPath := filepath.Join(directoryPath, server.Config.AppendFilename+".manifest")
	manifestFile, err := os.OpenFile(manifestPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	aofFileName, seq, _, err := readManifestFile(manifestFile)
	if err != nil {
		return err
	}

	if seq == -1 {
		manifestText := fmt.Sprintf("file %s seq 1 type i\n", server.Config.AppendFilename+".1.incr.aof")
		manifestFile.Write([]byte(manifestText))
		aofFileName = server.Config.AppendFilename + ".1.incr.aof"
	}
	manifestFile.Close()

	aofFilePath := filepath.Join(directoryPath, aofFileName)
	aofFile, err := os.OpenFile(aofFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil
	}

	server.isAofLoading = true
	server.aof = aofFile
	server.readAofFile()
	server.isAofLoading = false
	return nil
}

func readManifestFile(file *os.File) (string, int, string, error) {
	reader := bufio.NewReader(file)
	data, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return "", -1, "", errors.New("Failed to read the manifest file" + err.Error())
		}
		return "", -1, "", nil
	}

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

	return aofFileName, seq, fileType, nil
}

func (server *Server) readAofFile() error {
	if server.aof == nil {
		return errors.New("AOF not present on server's configuration")
	}

	server.aof.Seek(0, 0)
	reader := bufio.NewReader(server.aof)
	session := NewSession(nil, reader, nil)
	session.IsAuthenticated = true

	for {
		respCommand, err := readRespCommand(reader)

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		server.HandleCommand(session, respCommand)
	}
	return nil
}

func writeToAof(file *os.File, command string) error {
	if file == nil {
		return errors.New("AOF file is not open")
	}

	_, err := file.WriteString(command)
	if err != nil {
		fmt.Println("Error writing to AOF:", err)
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}
