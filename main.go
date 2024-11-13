package main

import (
	"encoding/json"
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/guptarohit/asciigraph"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gosuri/uilive"
)

type Currencies map[string]CurrencyValue

func UnmarshalCurrencies(data []byte) (Currencies, error) {
	var r Currencies
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Currencies) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type CurrencyValue struct {
	LastTrade string `json:"last_trade"`
}

func getData(pair string) (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.exmo.com/v1/ticker", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	dataCurrencies, err := UnmarshalCurrencies(bodyText)
	if err != nil {
		return 0, err
	}
	ticker, ok := dataCurrencies[pair]
	if !ok {
		return 0, fmt.Errorf("no such currency %s", pair)
	}
	price, err := strconv.ParseFloat(ticker.LastTrade, 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

func menu() {
	fmt.Print("\033[H\033[2J")
	fmt.Println("1. BTC_USD")
	fmt.Println("2. LTC_USD")
	fmt.Println("3. ETH_USD\n")

	fmt.Println("Press 1-3 to change symbol, press q to exit")
}

func main() {

	writer := uilive.New()

	menu()

	symbols := map[rune]string{
		'1': "BTC_USD",
		'2': "LTC_USD",
		'3': "ETH_USD",
	}

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	var ticker *time.Ticker
	quit := make(chan bool)
	var pair string

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		if char == 'q' {
			return
		}

		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			ticker.Stop()
			menu()
			continue
		}

		if char == '1' || char == '2' || char == '3' {
			if ticker != nil {
				// Stop the previous ticker and wait for the goroutine to finish
				ticker.Stop()
			}
			fmt.Print("\033[H\033[2J")

			quit = make(chan bool)

			pair = symbols[char]

			ticker = time.NewTicker(1 * time.Second)
			var data []float64

			// Запускаем поток вывода
			writer.Start()

			// Закрываем поток вывода при выходе из функции
			go func() {
				for {
					select {
					case <-ticker.C:
						price, err := getData(pair)
						if err != nil {
							fmt.Printf("Error getting data: %v\n", err)
						}

						data = append(data, price)

						graph := asciigraph.Plot(data, asciigraph.Width(100), asciigraph.Height(10), asciigraph.SeriesColors(asciigraph.Red))

						fmt.Fprintf(writer, "%s: %.2f\n", pair, price)
						fmt.Fprintln(writer, graph)
					case <-quit:
						writer.Stop()
						break
					}

				}
			}()
		}
	}

}
