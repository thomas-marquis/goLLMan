package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/controller/server/components"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/yuin/goldmark"
)

const sessionID = "masession" // TODO: remove this

func getSession(store session.Store, ctx context.Context) (*session.Session, error) {
	sess, err := store.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			sess, err = store.NewSession(ctx, session.WithLimit(10), session.WithID(sessionID))
		} else {
			return nil, err
		}
	}
	return sess, nil
}

func sendToStream(c *gin.Context, comp templ.Component) {
	buff := new(bytes.Buffer)
	if err := comp.Render(c.Request.Context(), buff); err != nil {
		pkg.Logger.Printf("Error rendering component: %s\n", err)
		showError(c, err, "Internal error", "")
		return
	}
	c.SSEvent("message", buff.String())
}

func convertToHTML(content string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func showError(c *gin.Context, err error, title, message string, msgArgs ...any) {
	if err == nil {
		c.HTML(http.StatusOK, "", components.Toast(
			components.ToastLevelError, title, fmt.Sprintf(message, msgArgs...)))
		return
	}
	if message == "" {
		c.HTML(http.StatusOK, "", components.Toast(
			components.ToastLevelError, title, err.Error()))
		return
	}
	c.HTML(http.StatusOK, "", components.Toast(
		components.ToastLevelError, title, fmt.Sprintf("%s: %s", fmt.Sprintf(message, msgArgs...), err.Error())))
}

func showInfo(c *gin.Context, title, message string, msgArgs ...any) {
	c.HTML(http.StatusOK, "", components.Toast(
		components.ToastLevelInfo, title, fmt.Sprintf(message, msgArgs...)))
}

func showSuccess(c *gin.Context, title, message string, msgArgs ...any) {
	c.HTML(http.StatusOK, "", components.Toast(
		components.ToastLevelSuccess, title, fmt.Sprintf(message, msgArgs...)))
}
