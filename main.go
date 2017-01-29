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
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				
				commandArray := strings.Split(message.Text, " ")
				if commandArray[0] != "roll" {
					break
				}
				rand.Seed(time.Now().UnixNano())
				var replyString string
				isCheckTypeFlag := true
				if len(commandArray) == 4 && commandArray[2] == "vs"{
					selfForce, parseSelfErr := strconv.Atoi(commandArray[1])
					if parseSelfErr != nil {
						break
					}
					targetForce, parseTargetErr := strconv.Atoi(commandArray[3])
					if parseTargetErr != nil {
						break
					}
					commandArray[1] = "對抗"
					commandArray[2] = strconv.Itoa(50 + ((selfForce - targetForce) * 5))
					isCheckTypeFlag = false
				}
				
				if len(commandArray) == 2 {
					replyString,_ = parseDiceArray(commandArray[1])
					if replyString == ""{
						break
					}
				}else{
					number, parseErr := strconv.Atoi(commandArray[2])
					if parseErr != nil {
						break
					}
					if number >= 100 {
						replyString = replyString + " 自動成功"
					}else if number <= 0 {
						replyString = replyString + " 自動失敗"
					}else{
						dice := rand.Intn(100) + 1  
						replyString = "《" + commandArray[1] + "》1D100<=" + strconv.Itoa(number) + "→" + strconv.Itoa(dice)
						
						isCheckTypeFlag = isCheckTypeFlag && !(commandArray[1] == "san" && len(commandArray) > 3)
						if dice > number {
							if isCheckTypeFlag && dice > 95 {
								replyString = replyString + " 大失敗"
							}else{
								replyString = replyString + " 失敗"
							}
						} else {
							if isCheckTypeFlag && dice == 1 {
								replyString = replyString + " 大成功"
							}else if isCheckTypeFlag && dice <= number / 5 {
								replyString = replyString + " 特別成功"
							}else{
								replyString = replyString + " 成功"
							}
						}
					
						if commandArray[1] == "san" && len(commandArray) > 3 {
							detectArray := strings.Split(commandArray[3], "/")
							if len(detectArray) == 2 {
								replyString = replyString + "\n"
								var detectString string
								if dice > number {
									detectString = detectArray[1]
								} else {
									detectString = detectArray[0]
								}
								diceResultString,diceResultInt := parseDiceArray(detectString)
								replyString = replyString + diceResultString + "\n《目前san值》" + strconv.Itoa(number) + "→" + strconv.Itoa(number - diceResultInt)
							}
						}
					}
				}
				
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyString)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func parseDiceArray(diceArrayString string) (replyString string,sum int){
	replyString = diceArrayString + "→"
	diceArray := strings.Split(diceArrayString, "+")
	multiFlag := len(diceArray) > 1
	sum = 0
	for index, dice := range diceArray {

		diceForamt := strings.Split(dice, "D")
		if len(diceForamt) > 2 {
			replyString = ""
			return
		}
		if len(diceForamt) == 1 {
			number, parseErr := strconv.Atoi(dice)
			if parseErr != nil {
				replyString = ""
				return
			}
			replyString = replyString + dice
			sum = sum + number
		}else{
			diceNumber, parseDiceNumberErr := strconv.Atoi(diceForamt[0])
			if parseDiceNumberErr != nil {
				replyString = ""
				return
			}
			diceType, parseDiceTypeErr := strconv.Atoi(diceForamt[1])
			if parseDiceTypeErr != nil {
				replyString = ""
				return
			}
			if diceNumber > 1 {
				multiFlag = true
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
	if multiFlag {
		replyString = replyString + "→" + strconv.Itoa(sum)
	}
	return 
}
