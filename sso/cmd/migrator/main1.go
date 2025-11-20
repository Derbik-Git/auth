package main // утилита для миграции БД

import (
	"errors"
	"fmt"
	"os"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	// Драйвер для выполнения миграций
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Задаем пути к хранилищу и миграциям
	storagePath := "C:/Users/user/Desktop/gRPC/sso/storage/sso.db"
	migrationsPath := "C:/Users/user/Desktop/gRPC/sso/migrations"
	migrationsTable := "migrations"

	// Проверяем, существует ли файл базы данных, если нет - создаем его
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		file, err := os.Create(storagePath)
		if err != nil {
			panic(fmt.Sprintf("Failed to create storage file: %v", err))
		}
		defer file.Close() // Закрываем файл после создания
	}

	// Создаем новый объект миграции
	m, err := migrate.New("file://"+migrationsPath, fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable))
	if err != nil {
		panic(fmt.Sprintf("Failed to create migration instance: %v", err))
	}

	// Применяем миграции
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}
		panic(fmt.Sprintf("Failed to apply migrations: %v", err))
	}

	fmt.Println("Migrations applied successfully")
}

// Мигратор — это инструмент для управления изменениями структуры базы данных. Когда база данных меняется (например, добавляется новая таблица или изменяется столбец), миграторы позволяют автоматически обновлять базу данных без потери данных. Это удобно особенно на продакшен-сервере, потому что ручное изменение схемы базы данных может привести к ошибкам.
