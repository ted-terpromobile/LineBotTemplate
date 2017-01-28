// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"
	"math/rand"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			messageSwitch:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				
				commandArray := strings.Split(message.Text, " ")
				if commandArray[0] != "roll" {
					break
				}
				rand.Seed(time.Now().UnixNano())
				var replyString string
				if len(commandArray) == 2 {
					replyString = commandArray[1] + "→"
					diceArray := strings.Split(commandArray[1], "+")
					var sum int = 0
					for index, dice := range diceArray {

						diceForamt := strings.Split(dice, "D")
						if len(diceForamt) > 2 {
							break messageSwitch
						}
						if len(diceForamt) == 1 {
							number, parseErr := strconv.Atoi(dice)
							if parseErr != nil {
								break messageSwitch
							}
							replyString = replyString + dice
							sum = sum + number
						}else{
							diceNumber, parseDiceNumberErr := strconv.Atoi(diceForamt[0])
							if parseDiceNumberErr != nil {
								break messageSwitch
							}
							diceType, parseDiceTypeErr := strconv.Atoi(diceForamt[1])
							if parseDiceTypeErr != nil {
								break messageSwitch
							}
							if diceNumber > 1 {
								replyString = replyString + "("
							}
							for i :=0 ; i < diceNumber ; i++ {
								diceEachResult := rand.Intn(diceType) + 1
								sum = sum + diceEachResult
								replyString = replyString + strconv.Itoa(diceEachResult)
								if i != (diceNumber - 1) {
									replyString = replyString + "+"
								}
							}
							if diceNumber > 1 {
								replyString = replyString + ")"
							}
						}
						
						if index != (len(diceArray) - 1) {
							replyString = replyString + "+"
						}
					}
					if len(diceArray) > 1 {
						replyString = replyString + "→"
					}
					replyString = replyString + strconv.Itoa(sum)
				}else{
				
					number, parseErr := strconv.Atoi(commandArray[2])
					if parseErr != nil {
						break
					}

					dice := rand.Intn(100) + 1  
					replyString = "《" + commandArray[1] + "》1D100<=" + strconv.Itoa(number) + "→" + strconv.Itoa(dice)

					if dice > number {
						replyString = replyString + " 失敗"
					} else {
						replyString = replyString + " 成功"
					}
				}
				
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyString)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}
