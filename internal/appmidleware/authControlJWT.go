package appmidleware

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/RangoCoder/foodApi/internal/env"
	st "github.com/RangoCoder/foodApi/internal/structs"
	"github.com/golang-jwt/jwt/v5"
)

/*
проверяем валидность входящего токена
если токен валиден - обновим и вернем свежий
если невалиден вернем пустую строку
*/
func ValidateJWT(key string) (string, error) {
	token, err := ValidateAcessToken(key)
	if err != nil {
		return "", errors.New("wrong token")
	}
	return token, nil
}

/*
проверка токена пользователя на валидность
на вход подаем accessToken
проверяем его и достаем payload
на основе полученных данных генерируем новый accessToken (строку)
если срок жизни токена уже вышел возвращаем пустую строку
*/
func ValidateAcessToken(tokenString string) (string, error) {
	var tokenClaims st.TokenClaims
	incomeClaims, err := GetClaimsFromJWT(tokenString)
	if err != nil { //
		return "", err
	}

	now := uint64(time.Now().Unix())              // текущее время
	sheeft := now - incomeClaims.Iat              // смотрим смещение
	jwtLiveTime := env.GetEnvVar("JWT_LIVE_TIME") // максимальное допустимое смещение в секундах (столько живет токен)
	maxsheeft, err := strconv.ParseUint(jwtLiveTime, 10, 64)
	if err != nil { //
		return "", err
	}
	if sheeft >= maxsheeft { // токен просрочен - отдаем  с сигналом о необходимости использовать токен для обновления
		return "needrefresh", nil
	}
	//токен еще живой - значит надо его обновить чтобы продлить ему жизнь
	tokenClaims.Sub = incomeClaims.Sub
	tokenClaims.Iat = now
	tokenClaims.Root = incomeClaims.Root
	// fmt.Printf("/n tokenClaims = %v", tokenClaims)

	accessToken, err := MakeAccessJWT(tokenClaims) //создаем новый актуальный JWT чтобы отдать пользователю
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func ExponentialConvert(expoNum any) (uint64, error) {
	expoStr := fmt.Sprint(expoNum) //сделаем строкой
	//проверим похожесть на экспоненциальное число
	separator := strings.Split(expoStr, "e+")             //разделим
	mantissa, err := strconv.ParseFloat(separator[0], 64) //возмем мантиссу
	if err != nil {                                       // ... handle error
		// fmt.Println("/n mantissa not geted")
		return 0.0, err
	}

	n, err := strconv.ParseUint(separator[1], 10, 64) // возмем степень
	if err != nil {                                   // ... handle error
		// fmt.Println("/n exp not geted")
		return uint64(0), err
	}
	var i, a, expresult uint64
	a = 10
	expresult = 1 // результат возведения в степень
	for i = 0; i < n; i++ {
		expresult *= a
	}
	// fmt.Printf("/n mantissa=%v/n ", mantissa)
	// fmt.Printf("/n expresult=%v/n ", expresult)
	finalRes := float64(mantissa) * float64(expresult)
	// fmt.Printf("/n finalRes=%v/n ", finalRes)
	return uint64(finalRes), nil
}

// разбераем токенна запчасти чтобы проверить
func GetClaimsFromJWT(tokenString string) (st.TokenClaims, error) {
	var err error
	var getedTokenClaims st.TokenClaims
	secretKeyString := env.GetEnvVar("JWT_SECRET_KEY")
	hmacSampleSecret := []byte(secretKeyString)

	// fmt.Printf("\n ValidateAcessToken %v", tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		log.Fatal(err)
	}

	// проверяем токен на валидность
	if !token.Valid { // если не валиден - прекращаем
		// fmt.Println("/n token NOT valid ")
		return st.TokenClaims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// fmt.Println(err)
		log.Fatal(err)
	}

	// fmt.Printf("/n claims= %v \n", claims)
	// преобразуем в uint
	// ид
	id, err := strconv.ParseUint(fmt.Sprint(claims["sub"]), 10, 64) // string to uint64
	if err != nil {                                                 // ... handle error
		// fmt.Println("/n id not geted")
		return st.TokenClaims{}, err
	}
	// fmt.Printf("/n id geted = %v", id)

	// время создания токена в экспоненциалном числе - тоже преобразуем к нормальному uint64
	createdTime, err := ExponentialConvert(claims["iat"])
	if err != nil { // ... handle error
		// fmt.Println("/n createdTime not geted ")
		return st.TokenClaims{}, err
	}
	// fmt.Printf("/n JWT createdTime geted = %v", createdTime)
	// ид
	root, err := strconv.ParseUint(fmt.Sprint(claims["root"]), 10, 64) // string to uint64
	if err != nil {                                                    // ... handle error
		// fmt.Println("/n id not geted")
		return st.TokenClaims{}, err
	}

	getedTokenClaims.Iat = createdTime
	getedTokenClaims.Sub = id
	getedTokenClaims.Root = root

	return getedTokenClaims, err
}

// создаем AccessJWT
func MakeAccessJWT(claims st.TokenClaims) (string, error) {
	fmt.Println("MakeAccessJWT claims ", claims)
	secretKeyString := env.GetEnvVar("JWT_SECRET_KEY")
	hmacSampleSecret := []byte(secretKeyString)
	nowTimeUnix := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  claims.Sub,
		"iat":  nowTimeUnix,
		"root": claims.Root,
	})

	newAccessToken, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		return "", err
	}
	return newAccessToken, nil
}

// func СomparisonIdByJWT(uid uint64, mytoken string) (bool, error) {
// 	var err error
// 	return false, err
// }
