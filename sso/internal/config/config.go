package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC        GRPCConfig    `yaml:"grpc"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	path := fetchConfigPath() // полученный путь к файлу конфигурации суём в переменную
	if path == "" {           // если путь пустой, то выпадаем с ошибкой
		panic("Config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) { // проверка на существование файла по извлечённому пути
		panic("Config file does not exist: " + path)
	}

	var config Config // переменная для обозначения структуры, для записи туда

	if err := cleanenv.ReadConfig(path, &config); err != nil { // чтение пути и запись в структуру
		panic("Error reading config file: " + err.Error())
	}
	/*
		// Обработка Duration
		if err := parseDuration(&config.TokenTTL); err != nil {
			panic(fmt.Sprintf("Error parsing token_ttl: %v", err))
		}
		if err := parseDuration(&config.GRPC.Timeout); err != nil {
			panic(fmt.Sprintf("Error parsing grpc timeout: %v", err))
		}
	*/
	return &config
}

// эта функция нужна для того, что бы извлекать флаг и парсить его, а после, функция выше использует эту функцию и достаёт из распаршеного фалага путь к конфигу
func fetchConfigPath() string {
	var result string

	//        переменная |  тег  |   знач. по умолч."" |  и в конце просто информативная строка
	flag.StringVar(&result, "config", "", "Path to config file") // достаём flag и записываем в переменную result
	flag.Parse()                                                 // парсим флаг (это обязательно) что бы работать с ним в коде

	if result == "" { // если переменная пустая то достаём значение из окружения по тегу и записываем в переменную
		result = os.Getenv("CONFIG_PATH") // достаём переменную из окружения (предварительно нужно установить переменную окружения через терминал ОС)
	}

	return result
}

/*
func parseDuration(d *time.Duration) error {
	dur, err := time.ParseDuration(d.String())
	if err != nil {
		return err
	}
	*d = dur
	return nil
}
*/
