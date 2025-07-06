package in_memory_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"
	"testing"
)

func Test_InMemorySessionStore_SaveAndGetByID(t *testing.T) {
	// Given
	ctx := context.TODO()
	store := in_memory.NewSessionStore()

	sess := session.New(session.WithID("test-session"))

	// When & Then
	err := store.Save(ctx, sess)
	assert.NoError(t, err)

	retrievedSess, err := store.GetByID(ctx, "test-session")
	assert.NoError(t, err)
	assert.Equal(t, sess.ID(), retrievedSess.ID())
}

func Test_InMemorySessionStore_GetByID_NotFound(t *testing.T) {
	// Given
	ctx := context.TODO()
	store := in_memory.NewSessionStore()

	// When
	retrievedSess, err := store.GetByID(ctx, "non-existent-session")

	// Then
	assert.Error(t, err)
	assert.Equal(t, session.ErrSessionNotFound, err)
	assert.Nil(t, retrievedSess)
}
