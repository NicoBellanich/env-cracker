package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

type EmbeddedFile struct {
	Filename string
	Ext      string
	Content  []byte
}

func main() {
	filePath := "sample.env" // archivo binario real

	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	// Separar por marcador `**%%DOCU`
	chunks := bytes.Split(data, []byte("**%%DOCU"))

	for i, chunk := range chunks {
		if i == 0 {
			continue // omitir cualquier header
		}

		metaAndData := extractMetadataAndData(chunk)
		if metaAndData != nil {
			if err := os.WriteFile(metaAndData.Filename, metaAndData.Content, 0644); err != nil {
				fmt.Printf("Error writing file %s: %v\n", metaAndData.Filename, err)
			} else {
				fmt.Println("Archivo extraído:", metaAndData.Filename)
			}
		}
	}
}

func extractMetadataAndData(chunk []byte) *EmbeddedFile {
	reader := bufio.NewReader(bytes.NewReader(chunk))
	var filename string
	var ext string

	// Buscar FILENAME y EXT en las primeras líneas
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineStr := string(line)

		if strings.HasPrefix(lineStr, "FILENAME/") {
			filename = strings.TrimPrefix(lineStr, "FILENAME/")
		}
		if strings.HasPrefix(lineStr, "EXT/") {
			ext = strings.TrimPrefix(lineStr, "EXT/")
		}

		if strings.HasPrefix(lineStr, "<?xml") || strings.Contains(lineStr, "RIFF") || strings.Contains(lineStr, "\xFF\xD8\xFF") {
			// Comienzo del contenido
			break
		}
	}

	// Si no hay nombre de archivo, saltear
	if filename == "" {
		return nil
	}

	// Usamos el resto del chunk como contenido
	contentIndex := bytes.Index(chunk, []byte("<?xml"))
	if contentIndex == -1 {
		contentIndex = bytes.Index(chunk, []byte("RIFF"))
	}
	if contentIndex == -1 {
		contentIndex = bytes.Index(chunk, []byte("\xFF\xD8\xFF")) // posible JPG
	}
	if contentIndex == -1 {
		return nil
	}

	content := chunk[contentIndex:]

	return &EmbeddedFile{
		Filename: filename,
		Ext:      ext,
		Content:  content,
	}
}
