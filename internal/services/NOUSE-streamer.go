package services

import (
	"fmt"
	"math/rand"
	"time"
)

func Streamer(data, status, exit chan int) {
	for {
		select {
		case <-data: // когда пришли какие-то данные - покажем их
			fmt.Println(data)
		case <-status: // пришли данные о новом статусе
			fmt.Printf("new status %v", status)
		case <-exit:
			fmt.Println("exit") //выходим и закрываем каналы
			return
		default:
			fmt.Println("wait....")
			time.Sleep(50 * time.Millisecond)
		}
	}
}

/*функция стримит в эфир данные о валютной паре*/
func TryStream() {
	data := make(chan int)   // канал с данными
	status := make(chan int) ///канал со статусом
	exit := make(chan int)   // сигнальный канал для выхода и закрытия каналов

	go func() { // асинхронно
		for { // в бескончном цикле
			gettedData := rand.Intn(100)      //считываем некоторые данные (например из базы или апи или еще откуда, да хоть из компьютерного зрения)
			data <- gettedData                //толкаем в канал с данными
			time.Sleep(50 * time.Millisecond) // немного отдыхаем
		}
	}()

	go func() { // в отдельной горутине даем команду на смену статуса
		time.Sleep(1 * time.Second)
		status <- 1
	}()

	go func() { // в отдельной горутине даем время работы до отключения
		time.Sleep(2 * time.Second)
		exit <- 1
	}()
	Streamer(data, status, exit) // запускаем стример в который регулярно отправляются данные
}
