package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"GRPC/sso/internal/app"
	"GRPC/sso/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//TODO: инициализировать объект конфига
	cfg := config.MustLoad()

	//TODO: инициализировать логгер
	log := setupLogger(cfg.Env)

	log.Info("staring application", slog.Any("config", cfg))

	application := app.NewAppMain(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustRun() // мы говорим что из этой функции, переменную(application), которой мы определили выше, перейди в элемент структуры GRPCSrv, тип которого *grpcapp.App, и вызови функцию MustRun

	//TODO: инициализировать приложение(код для запуска приложения)

	//TODO: запустить gRPC сервер приложения

	// простенький graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT) // graceful shutdown будет вызываться только в случае получения сигнала SIGTERM или SIGINT. То есть kill или ctrl+c

	sign := <-stop // Оператор <- используется для чтения из канала. Программа будет ждать здесь (блокировать выполнение), пока не будет получен сигнал. Как только сигнал поступит, он будет присвоен переменной sign.

	log.Info("stopping aplication", slog.String("signal", sign.String())) //  sign.String(): Этот метод преобразует полученный сигнал в строковое представление (например, "SIGINT" или "SIGTERM"), что делает логи более читаемыми.

	application.GRPCSrv.Stop()

	log.Info("application stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New( // В этом месте происходит сравнение: если env (который равен "local") совпадает с envLocal (который тоже равен "local"), то выполняется блок кода внутри этого case. смотри константы выше
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log

}
