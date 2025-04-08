package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"html"
	"os"
	"strings"
)

// DictionaryEntry represents a single entry in the Apple Dictionary XML format
type DictionaryEntry struct {
	XMLName xml.Name `xml:"d:entry"`
	ID      string   `xml:"id,attr"`
	Title   string   `xml:"d:title,attr"`
	Index   []Index  `xml:"d:index"`
	Content string   `xml:",innerxml"`
}

// Index represents the index element in the dictionary
type Index struct {
	XMLName xml.Name `xml:"d:index"`
	Value   string   `xml:"d:value,attr"`
}

// Dictionary represents the root element of the Apple Dictionary XML
type Dictionary struct {
	XMLName    xml.Name         `xml:"d:dictionary"`
	XMLNS      string          `xml:"xmlns,attr"`
	XMLNSD     string          `xml:"xmlns:d,attr"`
	Entries    []DictionaryEntry `xml:"d:entry"`
}

// sanitizeID creates a safe ID by replacing special characters
func sanitizeID(s string) string {
	// First escape HTML entities
	s = html.EscapeString(s)
	// Then replace remaining problematic characters
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "&", "_and_")
	s = strings.ReplaceAll(s, "/", "_slash_")
	s = strings.ReplaceAll(s, "\\", "_backslash_")
	s = strings.ReplaceAll(s, ":", "_colon_")
	s = strings.ReplaceAll(s, ";", "_semicolon_")
	s = strings.ReplaceAll(s, "?", "_question_")
	s = strings.ReplaceAll(s, "!", "_exclamation_")
	s = strings.ReplaceAll(s, "@", "_at_")
	s = strings.ReplaceAll(s, "#", "_hash_")
	s = strings.ReplaceAll(s, "$", "_dollar_")
	s = strings.ReplaceAll(s, "%", "_percent_")
	s = strings.ReplaceAll(s, "^", "_caret_")
	s = strings.ReplaceAll(s, "*", "_star_")
	s = strings.ReplaceAll(s, "(", "_lparen_")
	s = strings.ReplaceAll(s, ")", "_rparen_")
	s = strings.ReplaceAll(s, "+", "_plus_")
	s = strings.ReplaceAll(s, "=", "_equals_")
	s = strings.ReplaceAll(s, "[", "_lbracket_")
	s = strings.ReplaceAll(s, "]", "_rbracket_")
	s = strings.ReplaceAll(s, "{", "_lbrace_")
	s = strings.ReplaceAll(s, "}", "_rbrace_")
	s = strings.ReplaceAll(s, "|", "_pipe_")
	s = strings.ReplaceAll(s, "<", "_lt_")
	s = strings.ReplaceAll(s, ">", "_gt_")
	s = strings.ReplaceAll(s, ",", "_comma_")
	s = strings.ReplaceAll(s, ".", "_dot_")
	s = strings.ReplaceAll(s, "'", "_apos_")
	s = strings.ReplaceAll(s, "\"", "_quote_")
	return s
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <input.csv>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := "output.xml"

	// Open the CSV file
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}

	// Skip header row
	records = records[1:]

	// Create dictionary entries
	var entries []DictionaryEntry
	usedIDs := make(map[string]int) // Track used IDs and their counts

	for _, record := range records {
		if len(record) < 2 {
			continue
		}

		term := strings.TrimSpace(record[0])
		description := strings.TrimSpace(record[1])
		tag := ""
		if len(record) > 2 {
			tag = strings.TrimSpace(record[2])
		}

		// Create base ID by sanitizing the term
		baseID := sanitizeID(term)
		
		// Handle duplicate IDs
		count := usedIDs[baseID]
		usedIDs[baseID]++
		
		// If this is a duplicate, append a number to make it unique
		id := baseID
		if count > 0 {
			id = fmt.Sprintf("%s_%d", baseID, count)
		}

		// Create index with escaped value
		index := Index{Value: html.EscapeString(term)}

		// Create content with escaped values
		escapedTerm := html.EscapeString(term)
		escapedDescription := html.EscapeString(description)
		content := fmt.Sprintf("<h1>%s</h1>\n<div>\n%s\n</div>", escapedTerm, escapedDescription)
		if tag != "" {
			escapedTag := html.EscapeString(tag)
			content += fmt.Sprintf("\n<span class=\"tag\">%s</span>", escapedTag)
		}

		entry := DictionaryEntry{
			ID:      id,
			Title:   html.EscapeString(term),
			Index:   []Index{index},
			Content: content,
		}

		entries = append(entries, entry)
	}

	// Create dictionary
	dict := Dictionary{
		XMLNS:   "http://www.w3.org/1999/xhtml",
		XMLNSD:  "http://www.apple.com/DTDs/DictionaryService-1.0.rng",
		Entries: entries,
	}

	// Create output file
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	// Write XML header
	output.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")

	// Encode dictionary to XML
	encoder := xml.NewEncoder(output)
	encoder.Indent("", "\t")
	if err := encoder.Encode(dict); err != nil {
		fmt.Printf("Error encoding XML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
} 