// Package textsplit provides intelligent text splitting capabilities for TTS processing
package textsplit

import (
	"regexp"
	"strings"
)

const (
	MaxWordsPerChunk = 600 // Target words per chunk for optimal TTS generation
)

// TextChunk represents a piece of text with its order for processing
type TextChunk struct {
	Text  string
	Order int
}

// countWords counts the number of words in a text
func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// SplitText divides text into chunks of approximately MaxWordsPerChunk words,
// respecting sentence and paragraph boundaries
func SplitText(text string) []TextChunk {
	// Clean up the text
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}

	// If text is short enough, return as single chunk
	if countWords(text) <= MaxWordsPerChunk {
		return []TextChunk{{Text: text, Order: 0}}
	}

	chunks := []TextChunk{}
	order := 0

	// First split by paragraphs (double newlines have priority)
	paragraphs := regexp.MustCompile(`\n\s*\n`).Split(text, -1)

	currentChunk := ""
	currentWordCount := 0

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// Split paragraph into sentences
		sentences := splitIntoSentences(paragraph)

		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence == "" {
				continue
			}

			sentenceWordCount := countWords(sentence)

			// Check if adding this sentence would exceed the word limit
			if currentWordCount+sentenceWordCount > MaxWordsPerChunk && currentChunk != "" {
				// Save current chunk and start new one
				chunks = append(chunks, TextChunk{
					Text:  strings.TrimSpace(currentChunk),
					Order: order,
				})
				order++
				currentChunk = sentence
				currentWordCount = sentenceWordCount
			} else {
				// Add sentence to current chunk
				if currentChunk != "" {
					// Add space or newline depending on context
					if strings.HasSuffix(currentChunk, "\n") || strings.HasPrefix(sentence, "\n") {
						currentChunk += "\n" + sentence
					} else {
						currentChunk += " " + sentence
					}
				} else {
					currentChunk = sentence
				}
				currentWordCount += sentenceWordCount
			}
		}

		// Add paragraph break if we're continuing in the same chunk
		if currentChunk != "" && !strings.HasSuffix(currentChunk, "\n") {
			currentChunk += "\n"
		}
	}

	// Don't forget the last chunk
	if currentChunk != "" {
		chunks = append(chunks, TextChunk{
			Text:  strings.TrimSpace(currentChunk),
			Order: order,
		})
	}

	return chunks
}

// splitIntoSentences splits a paragraph into sentences using punctuation
func splitIntoSentences(paragraph string) []string {
	// Regex to split on sentence endings: . ! ? followed by space or end of string
	// But not on abbreviations like "Mr.", "Dr.", etc.
	sentenceRegex := regexp.MustCompile(`([.!?]+)(\s+|$)`)

	// Find all matches
	matches := sentenceRegex.FindAllStringIndex(paragraph, -1)
	if len(matches) == 0 {
		return []string{paragraph}
	}

	sentences := []string{}
	start := 0

	for _, match := range matches {
		end := match[1]
		sentence := paragraph[start:end]
		sentences = append(sentences, sentence)
		start = end
	}

	// Add remaining text if any
	if start < len(paragraph) {
		sentences = append(sentences, paragraph[start:])
	}

	return sentences
}


