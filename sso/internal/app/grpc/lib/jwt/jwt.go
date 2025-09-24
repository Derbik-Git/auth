package jwt

import (
	"GRPC/sso/internal/config/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) // 创建一个新的 JWT  签名方法 HS256

	claims := token.Claims.(jwt.MapClaims) // зполняем поля токена(claims), указываем тип мап jwt.MapClaims
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix // Время жизни токена, указывется при вызове функции
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString([]byte(app.Secret)) // подписываем токен полем секрета приложения, в виде массива байт
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

// можно покрыть эту функцию юнит тестами(потренироваться):
/*
package jwt

import (
 "GRPC/sso/internal/config/domain/models"
 "testing"
 "time"

 "github.com/golang-jwt/jwt/v5"
 "github.com/stretchr/testify/assert"
)

func TestNewToken(t *testing.T) {
 // Установим значения для теста
 user := models.User{
  ID:    "test-user-id",
  Email: "test@example.com",
 }
 app := models.App{
  ID:     "test-app-id",
  Secret: "test-secret",
 }
 duration := time.Minute

 // Вызовем функцию NewToken
 tokenString, err := NewToken(user, app, duration)

 // Проверим на наличие ошибок
 assert.NoError(t, err)

 // Проверим, что токен не пустой
 assert.NotEmpty(t, tokenString)

 // Попробуем распарсить токен
 token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
  // Проверяем метод подписи
  if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
   return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorUnverifiable)
  }
  return []byte(app.Secret), nil
 })

 // Проверяем на наличие ошибок при парсинге
 assert.NoError(t, err)

 // Проверяем, что токен валиден
 assert.True(t, token.Valid)

 // Проверяем поля токена (claims)
 claims, ok := token.Claims.(jwt.MapClaims)
 assert.True(t, ok)

 assert.Equal(t, user.ID, claims["uid"])
 assert.Equal(t, user.Email, claims["email"])
 assert.Equal(t, app.ID, claims["app_id"])

 // Проверяем время истечения токена
 expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
 assert.WithinDuration(t, time.Now().Add(duration), expirationTime, time.Second)
}

*/
