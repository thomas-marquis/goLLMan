package pkg

import "github.com/firebase/genkit/go/ai"

func ContentToText(content []*ai.Part) string {
	if len(content) != 1 || (len(content) >= 1 && content[0].Kind != ai.PartText) {
		Logger.Fatalf("WARNING unexpected message content: %v", content)
		return ""
	}
	return content[0].Text
}
