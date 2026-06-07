package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	DefaultPrimaryModel = "deepseek-v4-pro-260425"
	DefaultJudgeModel   = "doubao-seed-2-0-pro-260215"
	DefaultArkBaseURL   = "https://ark.cn-beijing.volces.com/api/v3"
)

// Config contains all runtime settings.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	LLM      LLMConfig
}

// AppConfig contains application-level settings.
type AppConfig struct {
	Env string
}

// ServerConfig contains Hertz listener settings.
type ServerConfig struct {
	Host string
	Port int
}

// Address returns the host:port string consumed by Hertz.
func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DatabaseConfig contains PostgreSQL settings.
type DatabaseConfig struct {
	URL string
}

// LLMConfig contains model provider settings. APIKey must never be logged.
type LLMConfig struct {
	Provider         string
	Model            string
	BackupModel      string
	JudgeModel       string
	APIKey           string
	BaseURL          string
	Temperature      float32
	JudgeTemperature float32
	MaxTokens        int
	JudgeMaxTokens   int
}

// Load reads local .env values and environment variables with safe defaults.
// Secret values are loaded but never logged.
func Load() (*Config, error) {
	_ = godotenv.Load()
	if envFile := strings.TrimSpace(os.Getenv("LLM_ENV_FILE")); envFile != "" {
		_ = godotenv.Load(envFile)
	}

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.env", "local")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.url", "")
	v.SetDefault("llm.provider", "ark")
	v.SetDefault("llm.model", DefaultPrimaryModel)
	v.SetDefault("llm.backup_model", "")
	v.SetDefault("llm.judge_model", DefaultJudgeModel)
	v.SetDefault("llm.api_key", "")
	v.SetDefault("llm.base_url", DefaultArkBaseURL)
	v.SetDefault("llm.temperature", 0.3)
	v.SetDefault("llm.judge_temperature", 0.0)
	v.SetDefault("llm.max_tokens", 2048)
	v.SetDefault("llm.judge_max_tokens", 1024)

	cfg := &Config{
		App: AppConfig{
			Env: v.GetString("app.env"),
		},
		Server: ServerConfig{
			Host: v.GetString("server.host"),
			Port: v.GetInt("server.port"),
		},
		Database: DatabaseConfig{
			URL: v.GetString("database.url"),
		},
		LLM: LLMConfig{
			Provider:         firstNonEmpty(v.GetString("llm.provider"), v.GetString("ark.provider"), "ark"),
			Model:            firstNonEmpty(v.GetString("llm.model"), v.GetString("ark.model"), DefaultPrimaryModel),
			BackupModel:      firstNonEmpty(v.GetString("llm.backup_model"), v.GetString("ark.backup_model")),
			JudgeModel:       firstNonEmpty(v.GetString("llm.judge_model"), v.GetString("judge.model"), DefaultJudgeModel),
			APIKey:           firstNonEmpty(v.GetString("llm.api_key"), v.GetString("ark.api_key"), v.GetString("openai.api_key")),
			BaseURL:          firstNonEmpty(v.GetString("llm.base_url"), v.GetString("ark.base_url"), DefaultArkBaseURL),
			Temperature:      float32(v.GetFloat64("llm.temperature")),
			JudgeTemperature: float32(v.GetFloat64("llm.judge_temperature")),
			MaxTokens:        v.GetInt("llm.max_tokens"),
			JudgeMaxTokens:   v.GetInt("llm.judge_max_tokens"),
		},
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("server port must be between 1 and 65535, got %d", cfg.Server.Port)
	}
	if cfg.LLM.MaxTokens <= 0 {
		return nil, fmt.Errorf("llm max_tokens must be greater than 0, got %d", cfg.LLM.MaxTokens)
	}
	if cfg.LLM.JudgeMaxTokens <= 0 {
		return nil, fmt.Errorf("llm judge_max_tokens must be greater than 0, got %d", cfg.LLM.JudgeMaxTokens)
	}
	if cfg.LLM.Temperature < 0 || cfg.LLM.Temperature > 1 {
		return nil, fmt.Errorf("llm temperature must be between 0 and 1, got %.2f", cfg.LLM.Temperature)
	}
	if cfg.LLM.JudgeTemperature < 0 || cfg.LLM.JudgeTemperature > 1 {
		return nil, fmt.Errorf("llm judge_temperature must be between 0 and 1, got %.2f", cfg.LLM.JudgeTemperature)
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(strings.Trim(value, `"'`))
		}
	}
	return ""
}
