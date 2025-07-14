package pkg

import "github.com/firebase/genkit/go/ai"

func ContentToText(content []*ai.Part) string {
	if len(content) != 1 || (len(content) >= 1 && content[0].Kind != ai.PartText) {
		Logger.Fatalf("WARNING unexpected message content: %v", content)
		return ""
	}
	return content[0].Text
}

func ContentFromText(text string) []*ai.Part {
	if text == "" {
		return nil
	}
	return []*ai.Part{
		{
			Kind: ai.PartText,
			Text: text,
		},
	}
}

func ClearMessageContext(msg []*ai.Message) {
	for _, m := range msg {
		if m.Role != ai.RoleUser {
			continue
		}

		var ctxPartIdx int = -1
		for i, part := range m.Content {
			if val, exists := part.Metadata["purpose"]; exists && val == "context" {
				ctxPartIdx = i
				break
			}
		}

		if ctxPartIdx != -1 {
			m.Content = m.Content[:ctxPartIdx]
		}
	}
}
