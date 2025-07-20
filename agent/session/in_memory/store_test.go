package in_memory_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/agent/session/in_memory"
	"testing"
)

func Test_InMemorySessionStore_NewSession_WithoutOptions(t *testing.T) {
	// Given
	store := in_memory.NewSessionStore()

	// When
	sess, err := store.NewSession(context.Background())

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.NotEmpty(t, sess.ID())

	found, err := store.GetByID(context.TODO(), sess.ID())
	assert.NotNil(t, found, "Session should exist in store")
	assert.NoError(t, err)
}

func Test_InMemorySessionStore_NewSession_WithOptions(t *testing.T) {
	// Given
	store := in_memory.NewSessionStore()

	// When
	sess, err := store.NewSession(context.Background(), session.WithLimit(33), session.WithID("toto"))

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 33, sess.Limit())
	assert.Equal(t, "toto", sess.ID())

	found, err := store.GetByID(context.TODO(), sess.ID())
	assert.NotNil(t, found, "Session should exist in store")
	assert.NoError(t, err)
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
