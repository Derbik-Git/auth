package suite

import (
	"GRPC/sso/internal/config"
	"context"
	"net"
	"strconv"
	"testing"

	ssov1 "github.com/Derbik-Git/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// тут будет код, связанный с работой над тестами

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const (
	grpcHost = "localhost"
)

// Эта функция New является конструктором контекста и тестового набора (Suite) для организации параллельных тестов и управления жизненным циклом тестирования в Go
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	// Если потребуется запускать тесты автоматически на гитхабе, то можно поступить следующим образом:
	/*
		path := "CONFIG_PATH"
		if v := os.Getenv(path); v != {
		return v
		}
	*/

	cfg := config.MustLoadByPath("../config/local.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout) // корневой контекст, с дед лайном на задачу, дедлайн указан в файле конфигурации

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	// DialContext - это функция для создлания нового gRPC клиента, по указанный параметрам(входящим)
	cc, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg), // адрес и тип среды, на котором будет работать клиент
		grpc.WithTransportCredentials(insecure.NewCredentials())) // опция, которая указывает что клиент будет использовать некодированные транспортные креденциалы(проще говоря данные без шифрования) | не ркуомендуется использовать в продакшене или произодственном окружении
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc), // создание клиента для работы с сервисом(метод из протофайла)
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port)) // функция, для объеденения порта(строки) и хоста(перевод из числа в строку), получается "host:port"
}
