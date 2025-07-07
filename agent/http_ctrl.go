package agent

import (
	"context"
	"github.com/thomas-marquis/goLLMan/pkg"
	"github.com/thomas-marquis/goLLMan/web"
	"net/http"

	genkit_core "github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	genkit_server "github.com/firebase/genkit/go/plugins/server"
)

type httpController struct {
	flow *genkit_core.Flow[ChatbotInput, string, struct{}]
	cfg  Config
}

func NewHTTPController(cfg Config, flow *genkit_core.Flow[ChatbotInput, string, struct{}]) *httpController {
	return &httpController{flow, cfg}
}

func (c *httpController) Run() error {
	ctx := context.Background()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", genkit.Handler(c.flow))
	mux.HandleFunc("GET /", web.IndexHandler())
	pkg.Logger.Fatal(genkit_server.Start(ctx, "127.0.0.1:3400", mux))
	return nil
}
