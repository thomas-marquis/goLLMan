package session_test

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/pkg"
	"testing"
)

func Test_Session_AddMessage_ShouldRemoveOldestMessageWhenLimitIsReached(t *testing.T) {
	// Given
	message := []*ai.Message{
		ai.NewMessage(ai.RoleSystem, nil, pkg.ContentFromText("system prompt: should not be removed")...),
		ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText("user message 1")...),
		ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText("assistant response 1")...),
		ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText("user message 2")...),
		ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText("assistant response 2")...),
		ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText("user message 3")...),
		ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText("assistant response 3")...),
	}

	sess := session.New(session.WithLimit(2))

	// When
	for _, msg := range message {
		assert.NoError(t, sess.AddMessage(msg))
	}

	// Then
	res, err := sess.GetMessages()
	assert.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Equal(t, "system prompt: should not be removed", pkg.ContentToText(res[0].Content))
	assert.Equal(t, "user message 3", pkg.ContentToText(res[1].Content))
	assert.Equal(t, "assistant response 3", pkg.ContentToText(res[2].Content))
}

func Test_Session_AddMessage_ShouldRemoveOldestMessageWithoutSystemMessage(t *testing.T) {
	// Given
	message := []*ai.Message{
		ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText("user message 1")...),
		ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText("assistant response 1")...),
		ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText("user message 2")...),
		ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText("assistant response 2")...),
		ai.NewMessage(ai.RoleUser, nil, pkg.ContentFromText("user message 3")...),
		ai.NewMessage(ai.RoleModel, nil, pkg.ContentFromText("assistant response 3")...),
	}

	sess := session.New(session.WithLimit(2))

	// When
	for _, msg := range message {
		assert.NoError(t, sess.AddMessage(msg))
	}

	// Then
	res, err := sess.GetMessages()
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "user message 3", pkg.ContentToText(res[0].Content))
	assert.Equal(t, "assistant response 3", pkg.ContentToText(res[1].Content))
}
