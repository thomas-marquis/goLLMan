package agent

import (
	"context"
	"net/http"

	genkit_core "github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	genkit_server "github.com/firebase/genkit/go/plugins/server"
)

type httpController struct {
	flow *genkit_core.Flow[string, string, struct{}]
}

func NewHTTPController(flow *genkit_core.Flow[string, string, struct{}]) *httpController {
	return &httpController{flow}
}

func (c *httpController) Run() error {
	ctx := context.Background()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", genkit.Handler(c.flow))
	logger.Fatal(genkit_server.Start(ctx, "127.0.0.1:3400", mux))
	return nil
}
