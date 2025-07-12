package server

import (
	"fmt"
	genkit_core "github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/thomas-marquis/goLLMan/agent"
	"github.com/thomas-marquis/goLLMan/agent/session"
	"github.com/thomas-marquis/goLLMan/controller/server/gintemplrenderer"
	"github.com/thomas-marquis/goLLMan/pkg"
	"log"
)

type Server struct {
	port         string
	host         string
	flow         *genkit_core.Flow[agent.ChatbotInput, string, struct{}]
	cfg          agent.Config
	sessionStore session.Store
	router       *gin.Engine
}

func New(
	cfg agent.Config,
	flow *genkit_core.Flow[agent.ChatbotInput, string, struct{}],
	sessionStore session.Store,
	g *genkit.Genkit,
) *Server {
	router := gin.Default()
	router.Static("/static", "./static")
	router.StaticFile("/favicon.ico", "./static/img/favicon.ico")

	stream := &eventStream{
		MessageBySessionID: make(map[string]messagesChan),
		NewClients:         make(chan messagesChan),
		ClosedClients:      make(chan messagesChan),
		TotalClients:       make(map[messagesChan]struct{}),
	}

	go stream.listen()

	s := &Server{
		port:         "3400",
		host:         "127.0.0.1",
		flow:         flow,
		cfg:          cfg,
		router:       router,
		sessionStore: sessionStore,
	}

	router.SetTrustedProxies(nil)
	ginHtmlRenderer := router.HTMLRender
	router.HTMLRender = &gintemplrenderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// TODO: finish implementing session management (kill idle or abusive users...)
	store := cookie.NewStore([]byte("secret")) // TODO: use a real secret here!
	router.Use(sessions.Sessions("chatsession", store))

	s.GetPageHandler(router)
	s.SSEMessagesHandler(router, sessionStore, stream)
	s.PostMessageHandler(router, sessionStore)
	s.FlowsHandlers(router, g)

	return s
}

func (s *Server) Run() error {
	pkg.Logger.Printf("Starting HTTP server on %s:%s", s.host, s.port)
	log.Fatal(s.router.Run(fmt.Sprintf(":%s", s.port)))
	return nil
}
