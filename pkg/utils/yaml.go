package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// YAMLDoc type for yaml files
	YAMLDoc = "yaml"
)

var (
	errHighDocumentCount = fmt.Errorf("document count was higher than expected count")
)

// LoadYAML loads a YAML file. Can return one or more IaC Documents.
// Besides reading in file data, its main purpose is to determine and store line number and filename metadata
func LoadYAML(filePath string) ([]*IacDocument, error) {
	iacDocumentList := make([]*IacDocument, 0)

	// First pass determines line number data
	{ // Limit the scope for Close()
		file, err := os.Open(filePath)
		if err != nil {
			return iacDocumentList, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		startLineNumber := 1
		currentLineNumber := 1
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "---") {
				// We've found the end-of-directives marker, so record results for the current document
				iacDocumentList = append(iacDocumentList, &IacDocument{
					Type:      YAMLDoc,
					StartLine: startLineNumber,
					EndLine:   currentLineNumber,
					FilePath:  filePath,
				})
				startLineNumber = currentLineNumber + 1
			}
			currentLineNumber++
		}

		// Add the very last entry
		iacDocumentList = append(iacDocumentList, &IacDocument{
			Type:      YAMLDoc,
			StartLine: startLineNumber,
			EndLine:   currentLineNumber,
			FilePath:  filePath,
		})

		if err = scanner.Err(); err != nil {
			return iacDocumentList, err
		}
	}

	// Second pass extracts all YAML documents and saves it in the document struct
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return iacDocumentList, err
	}

	dec := yaml.NewDecoder(bytes.NewReader(fileBytes))
	i := 0
	for {
		// each iteration extracts and marshals one yaml document
		var value interface{}
		err = dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return iacDocumentList, err
		}
		if i > (len(iacDocumentList) - 1) {
			return iacDocumentList, errHighDocumentCount
		}

		var documentBytes []byte
		documentBytes, err = yaml.Marshal(value)
		if err != nil {
			return iacDocumentList, err
		}
		iacDocumentList[i].Data = documentBytes
		i++
	}

	return iacDocumentList, nil
}
