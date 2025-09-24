package app

import (
	grpcapp "GRPC/sso/internal/app/grpc"
	auth1 "GRPC/sso/internal/services/auth"
	"GRPC/sso/internal/storage/sqlite"
	"log/slog"
	"time"
)

// ещё важно что в этой структуре я так понял мы указываем тип данных как указатель пакета.структуры, откуда мы берём функцию (то есть из какого файла)
type App struct {
	GRPCSrv *grpcapp.App // структура из файла генерации grpc сервера app.go |||||  Откуда берём функцию: Поскольку это указатель на структуру из другого пакета, вы можете использовать методы и функции этой структуры, не создавая новые экземпляры или копии.
}

func NewAppMain(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	// TODO: инициализировать хранилище (storage)
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	// TODO: инициализировать auth service (auth)
	authService := auth1.New(log, storage, storage, storage, tokenTTL)

	// TODO: инициализировать grpc server
	grpcApp := grpcapp.NewApp(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}

// для меня щас интересно как будет реализована всаимосвязь вот когда человек регается и ему присваивается TTL токен(как токен закрепляется за определённым человеком, хотя я думаю там всё проще чем я думаю) и как будет вообще выглядеть функция по созданию токена и функция для выдачи токена ну или как оно это там вообще будет
