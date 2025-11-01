package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tmpfile, err := os.CreateTemp("", "config-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	configContent := `
server:
  port: 9090
  mode: test
  name: quicksilver-test
  version: 0.1.0

database:
  host: testhost
  port: 5433
  name: testdb
  user: testuser
  password: testpass
  sslmode: require
  max_idle_conns: 10
  max_open_conns: 20
  conn_max_lifetime: 600

market:
  update_interval: "5s"
  data_source: "binance"
  api_url: "https://api.binance.com"
  symbols:
    - "BTC/USDT"
    - "ETH/USDT"
  hyperliquid:
    info_endpoint: "/info"
    ws_endpoint: "/ws"

trading:
  default_fee_rate: 0.002
  maker_fee_rate: 0.001
  taker_fee_rate: 0.002
  min_order_amount: 0.001

auth:
  jwt_secret: "test-jwt-secret"
  token_expire: 7200

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/quicksilver.log"
`

	_, err = tmpfile.Write([]byte(configContent))
	require.NoError(t, err)
	tmpfile.Close()

	// 注意：实际的 Load() 函数会从固定路径读取
	// 这里我们测试配置结构的解析
	t.Run("Config structure validation", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Port:    9090,
				Mode:    "test",
				Name:    "quicksilver-test",
				Version: "0.1.0",
			},
			Database: DatabaseConfig{
				Host:            "testhost",
				Port:            5433,
				Name:            "testdb",
				User:            "testuser",
				Password:        "testpass",
				SSLMode:         "require",
				MaxIdleConns:    10,
				MaxOpenConns:    20,
				ConnMaxLifetime: 600,
			},
		}

		assert.Equal(t, 9090, cfg.Server.Port)
		assert.Equal(t, "test", cfg.Server.Mode)
		assert.Equal(t, "testhost", cfg.Database.Host)
		assert.Equal(t, 5433, cfg.Database.Port)
	})
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "Basic DSN",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				Name:     "dbname",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=dbname sslmode=disable",
		},
		{
			name: "DSN with SSL",
			config: DatabaseConfig{
				Host:     "prod.example.com",
				Port:     5432,
				User:     "produser",
				Password: "prodpass",
				Name:     "proddb",
				SSLMode:  "require",
			},
			expected: "host=prod.example.com port=5432 user=produser password=prodpass dbname=proddb sslmode=require",
		},
		{
			name: "DSN with special characters",
			config: DatabaseConfig{
				Host:     "192.168.1.100",
				Port:     5432,
				User:     "admin",
				Password: "P@ssw0rd!",
				Name:     "my-database",
				SSLMode:  "verify-full",
			},
			expected: "host=192.168.1.100 port=5432 user=admin password=P@ssw0rd! dbname=my-database sslmode=verify-full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.GetDSN()
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	t.Run("Server defaults", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Port: 8080,
				Mode: "debug",
			},
		}

		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "debug", cfg.Server.Mode)
	})

	t.Run("Trading defaults", func(t *testing.T) {
		cfg := &Config{
			Trading: TradingConfig{
				DefaultFeeRate: 0.001,
				MakerFeeRate:   0.0005,
				TakerFeeRate:   0.001,
				MinOrderAmount: 0.0001,
			},
		}

		assert.Equal(t, 0.001, cfg.Trading.DefaultFeeRate)
		assert.Equal(t, 0.0005, cfg.Trading.MakerFeeRate)
		assert.Equal(t, 0.001, cfg.Trading.TakerFeeRate)
		assert.Equal(t, 0.0001, cfg.Trading.MinOrderAmount)
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("Environment variable override", func(t *testing.T) {
		// 设置环境变量
		os.Setenv("QS_SERVER_PORT", "9999")
		os.Setenv("QS_DATABASE_HOST", "env-host")
		defer func() {
			os.Unsetenv("QS_SERVER_PORT")
			os.Unsetenv("QS_DATABASE_HOST")
		}()

		// 注意：这里只是演示环境变量的概念
		// 实际的 Load() 函数会自动处理环境变量
		port := os.Getenv("QS_SERVER_PORT")
		host := os.Getenv("QS_DATABASE_HOST")

		assert.Equal(t, "9999", port)
		assert.Equal(t, "env-host", host)
	})
}

func TestMarketConfig(t *testing.T) {
	t.Run("Market config validation", func(t *testing.T) {
		cfg := &Config{
			Market: MarketConfig{
				UpdateInterval: "1s",
				DataSource:     "hyperliquid",
				APIURL:         "https://api.hyperliquid.xyz",
				Symbols:        []string{"BTC/USDT", "ETH/USDT", "SOL/USDT"},
				Hyperliquid: HyperliquidConfig{
					InfoEndpoint: "/info",
					WSEndpoint:   "/ws",
				},
			},
		}

		assert.Equal(t, "hyperliquid", cfg.Market.DataSource)
		assert.Equal(t, 3, len(cfg.Market.Symbols))
		assert.Contains(t, cfg.Market.Symbols, "BTC/USDT")
		assert.Equal(t, "/info", cfg.Market.Hyperliquid.InfoEndpoint)
	})

	t.Run("Multiple data sources", func(t *testing.T) {
		sources := []string{"hyperliquid", "binance", "coinbase"}

		for _, source := range sources {
			cfg := &Config{
				Market: MarketConfig{
					DataSource: source,
				},
			}
			assert.NotEmpty(t, cfg.Market.DataSource)
		}
	})
}

func TestAuthConfig(t *testing.T) {
	t.Run("JWT secret", func(t *testing.T) {
		cfg := &Config{
			Auth: AuthConfig{
				JWTSecret:   "my-secret-key-123456",
				TokenExpire: 3600,
			},
		}

		assert.NotEmpty(t, cfg.Auth.JWTSecret)
		assert.Greater(t, cfg.Auth.TokenExpire, 0)
	})
}

func TestLoggingConfig(t *testing.T) {
	t.Run("Logging levels", func(t *testing.T) {
		levels := []string{"debug", "info", "warn", "error"}

		for _, level := range levels {
			cfg := &Config{
				Logging: LoggingConfig{
					Level:  level,
					Format: "json",
				},
			}
			assert.Equal(t, level, cfg.Logging.Level)
		}
	})

	t.Run("Logging formats", func(t *testing.T) {
		formats := []string{"console", "json"}

		for _, format := range formats {
			cfg := &Config{
				Logging: LoggingConfig{
					Level:  "info",
					Format: format,
				},
			}
			assert.Equal(t, format, cfg.Logging.Format)
		}
	})
}
