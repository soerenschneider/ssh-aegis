package main

import (
	"bufio"
	"os"
)

type SshConfigWrapper struct {
	sshConfigFile string
}

func (s *SshConfigWrapper) GetConfig() ([]string, error) {
	inputFile, err := os.Open(s.sshConfigFile)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	var lines []string
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// WriteFile writes the modified contents back to the file
func (s *SshConfigWrapper) WriteConfig(data []string) error {
	outputFile, err := os.Create(s.sshConfigFile)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range data {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
