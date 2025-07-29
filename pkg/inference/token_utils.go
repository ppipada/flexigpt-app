package inference

import (
	"log/slog"
	"regexp"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
)

// FilterMessagesByTokenCount filters messages based on the maximum token count.
func FilterMessagesByTokenCount(
	messages []spec.ChatCompletionRequestMessage,
	maxTokenCount int,
) []spec.ChatCompletionRequestMessage {
	totalTokens := 0
	var filteredMessages []spec.ChatCompletionRequestMessage

	// Loop through the messages in reverse order (prioritizing the last element).
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		c := message.Content
		tokensInMessage := 0
		if c != nil {
			tokensInMessage = CountTokensInContent(*c)
		}

		// Check if adding this message will not exceed maxTokenCount
		// or if the filteredMessages slice is empty, then at least add this message.
		if totalTokens+tokensInMessage <= maxTokenCount || len(filteredMessages) == 0 {
			filteredMessages = append(filteredMessages, message)
			totalTokens += tokensInMessage

			// Always include at least one message, so if we've added one we can now enforce maxTokenCount.
			if totalTokens > maxTokenCount {
				break
			}
		} else {
			break
		}
	}

	if len(filteredMessages) < len(messages) {
		slog.Debug(
			"filtered messages are less than input",
			"messageCount",
			len(messages),
			"filteredCount",
			len(filteredMessages),
		)
	}

	// Reverse the filteredMessages slice to maintain the original order.
	for i, j := 0, len(filteredMessages)-1; i < j; i, j = i+1, j-1 {
		filteredMessages[i], filteredMessages[j] = filteredMessages[j], filteredMessages[i]
	}

	return filteredMessages
}

func CountTokensInContent(content string) int {
	// Regular expression to split the content into tokens based on common delimiters.
	// This includes whitespaces, brackets, arithmetic operators, and punctuation.
	tokenRegex := regexp.MustCompile(`[\s{}\[\]()+\-=*/<>,;:.!&|\\]+`)

	// Split the content into tokens based on the regex.
	tokens := tokenRegex.Split(content, -1)

	// Filter out empty strings and count the tokens.
	count := 0
	for _, token := range tokens {
		if token != "" {
			count++
		}
	}

	// Return the count of tokens.
	return count
}
