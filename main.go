package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	// Crear un directorio con hash único
	hash := generateHash()
	outputDir := fmt.Sprintf("extract-%s", hash)
	if err := os.Mkdir(outputDir, 0755); err != nil {
		panic(err)
	}
	fmt.Println("Directorio de extracción:", outputDir)

	// Separar por marcador `**%%DOCU`
	chunks := bytes.Split(data, []byte("**%%DOCU"))

	for i, chunk := range chunks {
		if i == 0 {
			continue // omitir cualquier header
		}

		metaAndData := extractMetadataAndData(chunk)
		if metaAndData != nil {
			outPath := filepath.Join(outputDir, metaAndData.Filename)
			if err := os.WriteFile(outPath, metaAndData.Content, 0644); err != nil {
				fmt.Printf("Error escribiendo %s: %v\n", outPath, err)
			} else {
				fmt.Println("Archivo extraído:", outPath)
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

	if filename == "" {
		return nil
	}

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

func generateHash() string {
	now := time.Now().String()
	h := sha1.New()
	h.Write([]byte(now))
	return fmt.Sprintf("%x", h.Sum(nil))[:8] // más corto y legible
}
