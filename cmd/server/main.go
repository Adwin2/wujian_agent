package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/adwin2/youthvital/api/handler"
	"github.com/adwin2/youthvital/internal/agent"
	"github.com/adwin2/youthvital/internal/config"
	"github.com/adwin2/youthvital/internal/repository"
	"github.com/adwin2/youthvital/internal/tool"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := repository.Open(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	tools, err := tool.NewRegistry().WithGraphTools(ctx)
	if err != nil {
		log.Fatalf("create graph tools: %v", err)
	}
	chatAgent := agent.NewPhase2ChatAgent(tools).WithAssessmentStore(db).WithAuditStore(db)
	if cfg.LLM.APIKey != "" {
		temperature := cfg.LLM.Temperature
		maxTokens := cfg.LLM.MaxTokens
		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:      cfg.LLM.APIKey,
			BaseURL:     cfg.LLM.BaseURL,
			Model:       cfg.LLM.Model,
			Temperature: &temperature,
			MaxTokens:   &maxTokens,
		})
		if err != nil {
			log.Fatalf("create ARK-compatible chat model: %v", err)
		}
		chatAgent, err = agent.NewEinoSupervisorChatAgent(ctx, chatModel, tools)
		if err != nil {
			log.Fatalf("create supervisor chat agent: %v", err)
		}
		chatAgent.WithAssessmentStore(db).WithAuditStore(db)
	}

	h := server.Default(server.WithHostPorts(cfg.Server.Address()))
	healthHandler := handler.NewHealthHandler(nil)
	if db != nil {
		healthHandler = handler.NewHealthHandler(db)
	}
	healthHandler.Register(h)

	v1 := h.Group("/v1")
	chatHandler := handler.NewChatHandler(chatAgent)
	chatHandler.Register(v1)

	log.Printf("YouthVital Phase 2 server listening on %s", cfg.Server.Address())
	h.Spin()
}
