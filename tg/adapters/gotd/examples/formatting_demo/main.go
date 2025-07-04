package main

import (
	"fmt"
	"github.com/bitia-ru/gotg/tg"
)

func main() {
	// This example demonstrates the formatted message functionality
	// It shows how to create different types of message chunks and combine them

	fmt.Println("=== Message Formatting Demo ===")
	
	// Example 1: Simple text with basic formatting
	fmt.Println("\n1. Basic formatting:")
	basicChunk := tg.Container(
		tg.Text("Hello, this is "),
		tg.Bold("bold text"),
		tg.Text(" and "),
		tg.Italic("italic text"),
		tg.Text("!"),
	)
	printChunkDemo(basicChunk)

	// Example 2: Code and preformatted text
	fmt.Println("\n2. Code formatting:")
	codeChunk := tg.Container(
		tg.Text("Here's some "),
		tg.Code("inline code"),
		tg.Text(" and a code block:\n"),
		tg.PreCode("func main() {\n    fmt.Println(\"Hello, world!\")\n}", "go"),
	)
	printChunkDemo(codeChunk)

	// Example 3: Complex formatting with links
	fmt.Println("\n3. Complex formatting:")
	complexChunk := tg.Container(
		tg.Bold("Important:"),
		tg.Text(" Visit our "),
		tg.Link("website", "https://example.com"),
		tg.Text(" for more info about "),
		tg.Spoiler("spoiler content"),
		tg.Text(" and "),
		tg.Strike("strikethrough"),
		tg.Text(" text."),
	)
	printChunkDemo(complexChunk)

	// Example 4: Nested containers
	fmt.Println("\n4. Nested formatting:")
	nestedChunk := tg.Container(
		tg.Text("This is a "),
		tg.Container(
			tg.Bold("nested"),
			tg.Text(" "),
			tg.Italic("container"),
		),
		tg.Text(" example."),
	)
	printChunkDemo(nestedChunk)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Note: This demo shows the structure. To actually send formatted messages,")
	fmt.Println("use peer.SendMessageFormatted(ctx, chunk) with proper Telegram credentials.")
}

// Helper function to demonstrate chunk structure
func printChunkDemo(chunk tg.MessageChunk) {
	options := chunk.ToStyledTextOptions()
	fmt.Printf("Chunk contains %d styled text options\n", len(options))
	
	// In a real application, you would use:
	// peer.SendMessageFormatted(ctx, chunk)
	// or
	// message.ReplyFormatted(ctx, tg, chunk)
}