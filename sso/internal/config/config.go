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

	return MustLoadByPath(path)
}

// !!!Мы могли юы просто логику этой функции перенести в MustLoad, но мы так делаем для возможно работы с тестами, в тестах мы будем вызывать MustloadByPath, так как она принимает обычный строковый аргумент и у нс проще говоря должно быть фейковое значение, то есть мы передаём тудf уже какую то создланную константу или что то типо того, а не достаём настоящий путь из флага, за счёт этого мы сможем в тестах вызвать на прямую функцию MustLoadByPath и передать туда путь к файлу конфигурации
func MustLoadByPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) { // проверка на существование файла по извлечённому пути
		panic("Config file does not exist: " + configPath)
	}

	var cfg Config // переменная для обозначения структуры, для записи туда

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil { // чтение пути и запись в структуру
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
	return &cfg
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
