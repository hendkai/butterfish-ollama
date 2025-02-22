package butterfish

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixCommandParse(t *testing.T) {
	str1 := `
Foo bar foo bar

> command arg1 arg2 "arg3 arg4" arg5 -v

Foo bar foo bar`

	cmd, err := fixCommandParse(str1)
	assert.Nil(t, err)
	assert.Equal(t, "command arg1 arg2 \"arg3 arg4\" arg5 -v", cmd)

	str2 := `
Foo bar foo bar

` + "```" + `
command arg1 arg2 "arg3 arg4" arg5 -v
` + "```" + `

Foo bar foo bar`

	cmd, err = fixCommandParse(str2)
	assert.Nil(t, err)
	assert.Equal(t, "command arg1 arg2 \"arg3 arg4\" arg5 -v", cmd)
}

// A golang test for ShellBuffer
func TestShellBuffer(t *testing.T) {
	buffer := NewShellBuffer()
	buffer.Write("hello")
	assert.Equal(t, "hello", buffer.String())
	buffer.Write(" world")
	assert.Equal(t, "hello world", buffer.String())
	buffer.Write("!")
	assert.Equal(t, "hello world!", buffer.String())
	buffer.Write("\x1b[D")
	assert.Equal(t, "hello world!", buffer.String())

	buffer = NewShellBuffer()
	buffer.Write("~/butterfish ᐅ")
	assert.Equal(t, "~/butterfish ᐅ", buffer.String())

	// test weird ansii escape sequences
	red := "\x1b[31m"
	buffer = NewShellBuffer()
	buffer.Write("foo")
	buffer.Write(red)
	buffer.Write("bar")
	assert.Equal(t, "foo"+red+"bar", buffer.String())

	// test shell control characters
	buffer = NewShellBuffer()
	buffer.Write(string([]byte{0x6c, 0x08, 0x6c, 0x73, 0x20}))
	assert.Equal(t, "ls ", buffer.String())

	// test left cursor, backspace, and then insertion
	buffer = NewShellBuffer()
	buffer.Write("hello world")
	buffer.Write("\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D")
	buffer.Write("foo   ")
	buffer.Write("\x08\x7f") // backspace
	assert.Equal(t, "hello foo world", buffer.String())
}

// function to test shell history using golang testing tools
func TestShellHistory(t *testing.T) {
	history := NewShellHistory()

	history.Append(historyTypePrompt, "prompt1")
	history.Append(historyTypeShellInput, "shell1")
	history.Append(historyTypeShellOutput, "output1")
	history.Append(historyTypeLLMOutput, "llm1")

	output := HistoryBlocksToString(history.GetLastNBytes(256, 512))
	assert.Equal(t, "prompt1\nshell1\noutput1\nllm1", output)

	history.Append(historyTypePrompt, "prompt2 xxxxxxxxxxxxxxxxxxxxxxxxxxxxx       xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx             xxxxxxxxxxxxxxxxxxxxxxxxxxxxx         xxxxxxxxxx         xxxxxxxxxxxxxxxxxxx               xxxxxxxxxxxxxxxxxxxxx")
	history.Append(historyTypeLLMOutput, "llm2")

	output = HistoryBlocksToString(history.GetLastNBytes(14, 512))
	assert.Equal(t, "llm2", output)

	history.Append(historyTypeLLMOutput, "more llm ᐅ")
	output = HistoryBlocksToString(history.GetLastNBytes(24, 512))
	assert.Equal(t, "llm2more llm ᐅ", output)
}

// A test case for incompleteAnsiSequence()
func TestIncompleteAnsiSequence(t *testing.T) {
	// incomplete sequence
	assert.True(t, incompleteAnsiSequence([]byte{0x1b, 0x5b, 0x30, 0x3b}))
	assert.True(t, incompleteAnsiSequence([]byte{0x20, 0x1b, 0x5b, 0x30, 0x3b}))
	// complete sequence
	assert.False(t, incompleteAnsiSequence([]byte{0x1b, 0x5b, 0x30, 0x3b, 0x31, 0x3b, 0x32, 0x6d, 0x1b, 0x5b, 0x30, 0x6d}))
	assert.False(t, incompleteAnsiSequence([]byte{0x20, 0x20, 0x1b, 0x5b, 0x30, 0x3b, 0x31, 0x3b, 0x32, 0x6d, 0x1b, 0x5b, 0x30, 0x6d}))
}

// Test for Ollama completions
func TestOllamaCompletion(t *testing.T) {
	ollama := NewOllama("test-token", "https://api.ollama.com")
	request := &util.CompletionRequest{
		Prompt:      "Hello, world!",
		Model:       "ollama-model",
		MaxTokens:   10,
		Temperature: 0.7,
	}

	response, err := ollama.Completion(request)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Completion)
}

// Test for Ollama embeddings
func TestOllamaEmbeddings(t *testing.T) {
	ollama := NewOllama("test-token", "https://api.ollama.com")
	input := []string{"Hello, world!", "How are you?"}

	embeddings, err := ollama.Embeddings(context.Background(), input, false)
	assert.Nil(t, err)
	assert.NotNil(t, embeddings)
	assert.Equal(t, len(input), len(embeddings))
	for _, embedding := range embeddings {
		assert.NotEmpty(t, embedding)
	}
}
