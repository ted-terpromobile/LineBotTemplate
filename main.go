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
	"math"
	
	"path/filepath"
// 	"io/ioutil"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client
var appBaseURL string
var downloadDir string

func main() {
	var err error
	
	appBaseURL = "https://lolilinebot.herokuapp.com"
	
	//
	downloadDir = filepath.Join(filepath.Dir(os.Args[0]), "saveData")
	_, err = os.Stat(downloadDir)
	if err != nil {
		err = os.Mkdir(downloadDir, 0777)
	}
	//
	
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	
	imageFileServer := http.FileServer(http.Dir("images"))
	http.HandleFunc("/images/", http.StripPrefix("/images/", imageFileServer).ServeHTTP)
	
// 	LogServer := http.FileServer(http.Dir(downloadDir))
// 	http.HandleFunc("/log/", http.StripPrefix("/log/", LogServer).ServeHTTP)
	
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

//
func saveText(text string, overWrite bool) (*os.File, error) {
	file, err := os.OpenFile(downloadDir + "/saveText", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil || overWrite {
		file, err = os.Create(downloadDir + "/saveText")
		if err != nil {
			return nil, err
		}
// 		err = file.Chmod(0777)
// 		if err != nil {
// 			return nil, err
// 		}
	}else{
		if text != "" {
			text = text + "\n"
		}
	}
	defer file.Close()
	
	_,err = file.WriteString(text)
	if err != nil {
		return nil, err
	}
	
	return file, nil
}

func loadText() (string, error) {
	file, err := os.Open(downloadDir + "/saveText")
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	readBytes := make([]byte, fileInfo.Size())
	_,err = file.Read(readBytes)
	if err != nil {
		return "", err
	}
	
	return string(readBytes), nil
}
//

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
		displayName := ""
		if event.Source.UserID != "" {
			profile, err := bot.GetProfile(event.Source.UserID).Do()
			if err == nil {
				displayName = profile.DisplayName
			}
		}	
		
		if event.Type == linebot.EventTypePostback {
				replyString := ""
				if event.Postback.Data == "自我介紹"{
					replyString = displayName + "您好^^，我是Ted的女兒。現在的工作是幫大家擲骰子!擲出壞數字也不可以怪我喔!"
				}
				if event.Postback.Data == "說明"{
					replyString = 	"《一般擲骰指令》\n" +
							"假設要骰3次20面骰然後數值再加1，就輸入：\n" +
							"roll 3D20+1\n" +
							"（roll 空格 骰子表示法）\n" +
							"《一般技能指令》\n" +
							"假設要骰觀察，觀察技能有60，就輸入：\n" +
							"roll 觀察 60\n" +
							"（roll 空格 技能名稱 空格 技能數值）\n" +
							"《SanCheck指令》\n" +
							"假設我san40，然後KP說1/1D8，就輸入：\n" +
							"roll san 40 1/1D8\n" +
							"（roll 空格 包含san字串的文字 空格 san的數值 空格 sancheck數值）\n" +
							"《對抗指令》\n" +
							"假設我要暴力開門，我力量10，門的抵抗力量5，就輸入：\n" +
							"roll 10 vs 5\n" +
							"（roll 空格 你的力量 空格 vs 空格 另一個抵抗力）"
				}
				if event.Postback.Data == "exit"{
					replyString = "掰掰!"
				}
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyString)).Do(); err != nil {
					log.Print(err)
				}
				if event.Postback.Data == "exit"{
					switch event.Source.Type {
						case linebot.EventSourceTypeGroup:
							bot.LeaveGroup(event.Source.GroupID).Do()
						case linebot.EventSourceTypeRoom:
							bot.LeaveRoom(event.Source.RoomID).Do()
					}
				}
		}
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {	
			case *linebot.TextMessage:
				var replyString string		
				
				// Line的連續空格會變 \u00a0
				for ; strings.Contains(message.Text, "\u00a0"); {
					message.Text = strings.Replace(message.Text, "\u00a0", " ",-1)
				}
				for ; strings.Contains(message.Text, "　"); {
					message.Text = strings.Replace(message.Text, "　", " ",-1)
				}
				for ; strings.Contains(message.Text, "  "); {
					message.Text = strings.Replace(message.Text, "  ", " ",-1)
				}
				//wordGame
				//wordGameLose := ""
				//wordGame,err := loadText()
				//wordGameWords := strings.Split(wordGame, "\n")
				//roomIDTimes := 0
				//for _, word := range wordGameWords {
				//	if strings.Contains(message.Text, word) && strings.Contains(message.Text, "\""){
				//		wordGameLose = word
				//	}
				//	if event.Source.RoomID == word{
				//		roomIDTimes++
				//	}
				//	if event.Source.RoomID == word && wordGameLose != ""{
				//		if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("輸了 禁詞:" + wordGameLose)).Do(); err != nil {
				//			log.Print(err)
				//		}
				//		return
				//	}
				//}
				//wordGameend
				
				commandArray := strings.Split(message.Text, " ")
				if strings.ToLower(commandArray[0]) != "roll" {
					
					if strings.Contains(strings.ToLower(commandArray[0]), "c") || strings.Contains(strings.ToLower(commandArray[0]), "p"){
						chooseArray := strings.Split(commandArray[0], "c")
						if len(chooseArray) != 2{
							chooseArray = strings.Split(commandArray[0], "p")
						}
						if len(chooseArray) != 2{
							return
						}
						number, parseErr := strconv.Atoi(chooseArray[0])
						if parseErr != nil {
							replyString = ""
							return
						}
						number2, parseErr2 := strconv.Atoi(chooseArray[1])
						if parseErr2 != nil {
							replyString = ""
							return
						}
						if number < number2 {
							return
						}
						
						numberArray := make([]string, number)
						for i := 0 ; i < number ; i++{
							numberArray[i] = strconv.Itoa(i+1)
						}
						
						rand.Seed(time.Now().UnixNano())
						for j := 0 ; j < (number-number2) ; j++{
							chosenPos := rand.Intn(len(numberArray))
							numberArray = append(numberArray[:chosenPos], numberArray[(chosenPos+1):]...)
						}
						
						if strings.Contains(strings.ToLower(commandArray[0]), "p"){
							for k:=0 ; k < len(numberArray) ; k++{
								chosenPos2 := rand.Intn(len(numberArray))
								temp := numberArray[k]
								numberArray[k] = numberArray[chosenPos2]
								numberArray[chosenPos2] = temp
							}
						}
						
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(strings.Join(numberArray, ", "))).Do(); err != nil {
							log.Print(err)
						}
						return
					}
					////wordGame
					//if commandArray[0] == "禁詞遊戲開始" {
					//	replySaved := "記錄錯誤"
					//	_,err := saveText(event.Source.RoomID,false)
					//	if err == nil{
					//		replySaved = "開始! 現有" + strconv.Itoa(len(wordGameWords) - (1+roomIDTimes)) + "位玩家"					
					//	}
					//	if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replySaved)).Do(); err != nil {
					//		log.Print(err)
					//	}
					//	return
					//}
					//if commandArray[0] == "禁詞" {
					//	replySaved := "記錄錯誤"
					//	if len(commandArray[1]) > 1 {
					//		_,err := saveText(commandArray[1],false)
					//		if err == nil{
					//			replySaved = "記錄成功"
					//		}
					//	}
					//	if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replySaved)).Do(); err != nil {
					//		log.Print(err)
					//	}
					//	return
					//}
					//if commandArray[0] == "禁詞遊戲準備" {
					//	replySaved := "記錄錯誤"
					//	_,err := saveText("",true)
					//	if err == nil{
					//		replySaved = "私密ㄌㄌ輸入 禁詞 (空格) (你選的禁詞) ，全部玩家輸入完後開始!"
					//	}
					//	if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replySaved)).Do(); err != nil {
					//		log.Print(err)
					//	}
					//	return
					//}
					//wordGame end
// 					if commandArray[0] == "new" {
// 						_,err := saveText(strings.Replace(message.Text, "new ", "", 1),true)
// 						replySaved := "saved"
// 						if err != nil{
// 							replySaved = err.Error()
// 						}
// 						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replySaved)).Do(); err != nil {
// 							log.Print(err)
// 						}
// 						return
// 					}
// 					if commandArray[0] == "save" {
// 						_,err := saveText(strings.Replace(message.Text, "save ", "", 1),false)
// 						replySaved := "saved"
// 						if err != nil{
// 							replySaved = err.Error()
// 						}
// 						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replySaved)).Do(); err != nil {
// 							log.Print(err)
// 						}
// 						return
// 					}
// 					if commandArray[0] == "load" {
// 						loadText,err := loadText()
// 						if err != nil {
// 							loadText = err.Error()
// 						}
// 						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(loadText)).Do(); err != nil {
// 							log.Print(err)
// 						}
// 						return
// 					}
					if commandArray[0] == "今日運勢" {
						runes := []rune(time.Now().Month().String() + event.Source.UserID)
						sum := 0
						for _, each := range runes {
							sum = sum * int(each) + time.Now().Day() * time.Now().Year()
						}
						
						replyLuck := displayName + "的今日運勢是 "
						switch {
						case sum % 20 == 0:
							replyLuck = replyLuck + "大凶"
						case sum % 20 < 5:
							replyLuck = replyLuck + "凶"
						case sum % 20 < 10:
							replyLuck = replyLuck + "中吉"
						case sum % 20 < 12:
							replyLuck = replyLuck + "大吉"
						default:
							replyLuck = replyLuck + "小吉"
						}
						
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyLuck)).Do(); err != nil {
							log.Print(err)
						}
						return
					}
						
					if commandArray[0] == "GM" {
						if event.Source.UserID != "Ue31a3821dcc6848bb9b9e6080cc584ba" {
							return
						}
						if event.Source.GroupID != "" || event.Source.RoomID != ""{
							return
						}
						replySaved := "記錄錯誤"
						_,err := saveText(commandArray[1],true)
						if err == nil{
							replySaved = "記錄成功"
						}
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replySaved)).Do(); err != nil {
							log.Print(err)
						}
						return
					}
					
					if commandArray[0] == "ㄌㄌ" || commandArray[0] == "莉亞"{
						//noDiceReplyString := "你說的話是什麼意思? 對不起，我聽不懂QAQ"
						//if strings.Contains(commandArray[1], "自我介紹"){
						//	noDiceReplyString = "大家好^^，我是Ted的女兒。現在的工作是幫大家擲骰子! 擲出壞數字也不可以怪我喔!"    						
						//}
						//if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(noDiceReplyString)).Do(); err != nil {
						//	log.Print(err)
						//}
						imageURL := appBaseURL + "/images/loli.jpg"
						template := linebot.NewButtonsTemplate(
							imageURL, "莉亞", displayName + "有什麼事嗎?",
							linebot.NewPostbackTemplateAction("你是誰?", "自我介紹", "你是誰?"),
							linebot.NewPostbackTemplateAction("怎麼擲骰子呢?", "說明","怎麼擲骰子呢?"),
							linebot.NewPostbackTemplateAction("辛苦了，去休息吧。", "exit","辛苦了，去休息吧。"),
						)
						if _, err := bot.ReplyMessage(
							event.ReplyToken,
							linebot.NewTemplateMessage(displayName + "有什麼事嗎?", template),
						).Do(); err != nil {
							return
						}
					}					
					break
				}
				rand.Seed(time.Now().UnixNano())
				isCheckTypeFlag := true
				if len(commandArray) >= 4 && strings.ToLower(commandArray[2]) == "vs"{
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
					if len(commandArray) == 5{
						commandArray[3] = commandArray[4]
					}else{
						commandArray[3] = "0"
					}
				}
				
				if len(commandArray) == 2 {
					replyString,_ = parseDiceArray(commandArray[1],false)
					if replyString == ""{
						break
					}
				}else{
					//忽略指令2段中的空格
					commandCopyArray := make([]string, len(commandArray))
					copy(commandCopyArray, commandArray)
					for i := 2 ; i < len(commandCopyArray) ; i++ {
						_, isString := strconv.Atoi(commandCopyArray[i])
						if isString != nil {
							commandArray[1] = commandArray[1] + " " + commandCopyArray[i]
						} else {
							commandArray[2]	= commandCopyArray[i]
							if len(commandArray) > 3 {
								if i+1 < len(commandCopyArray){
									commandArray[3]	= commandCopyArray[i+1]
								} else {
									commandArray[3] = "0"
								}
								for j := 4 ; j < len(commandArray) ; j++ {
									commandArray[j] = "0"
								}
							}
							break
						}
					}
					//
					number, parseErr := strconv.Atoi(commandArray[2])
					if parseErr != nil {
						replyString,_ = parseDiceArray(commandArray[2],false)
						if replyString == ""{
							break
						}
						if(!strings.Contains(strings.ToLower(commandArray[1]), "我們")){
							commandArray[1] = strings.Replace(commandArray[1], "我", displayName,-1)
						}
						replyString = "《" + commandArray[1] + "》" + replyString
					}else{
						if number >= 100 {
							replyString = replyString + " 自動成功"
						}else if number <= 0 {
							replyString = replyString + " 自動失敗"
						}else{
							plusDiceStr := ""
							if len(commandArray) > 3 {
								plusDiceStr = commandArray[3]
							}
							plusDice, plusDiceErr := strconv.Atoi(plusDiceStr)
							if plusDiceErr != nil{ //!isCheckTypeFlag ||
								plusDice = 0
							}

							dice := rand.Intn(100) + 1

							//GM
							loadData,loadErr := loadText()
							if loadErr == nil && loadData != ""{
								diceGM, GMErr := strconv.Atoi(loadData)
								if GMErr == nil && diceGM <= 100{
									dice = diceGM
								}
								saveText("",true)
							}
							//GMend
							if(!strings.Contains(strings.ToLower(commandArray[1]), "我們")){
								commandArray[1] = strings.Replace(commandArray[1], "我", displayName,-1)
							}
							replyString = "《" + commandArray[1] + "》1D100<=" + strconv.Itoa(number) + "→" + strconv.Itoa(dice)

							for i := 0.0; i < math.Abs(float64(plusDice)); i++ {
								diceTemp := rand.Intn(100) + 1  
								replyString = replyString + "\n1D100<=" + strconv.Itoa(number) + "→" + strconv.Itoa(diceTemp)
								if (plusDice > 0 && diceTemp < dice) || (plusDice < 0 && diceTemp > dice) {
									dice = diceTemp
								}
							}
							if plusDice != 0 {
								replyString = replyString + "\n"
							}

							//isCheckTypeFlag = isCheckTypeFlag && !strings.Contains(strings.ToLower(commandArray[1]), "san")
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
								}else if isCheckTypeFlag && dice <= number / 2 {
									replyString = replyString + " 較成功"
								}else{
									replyString = replyString + " 成功"
								}
							}

							if strings.Contains(strings.ToLower(commandArray[1]), "san") && len(commandArray) > 3 {
								detectArray := strings.Split(commandArray[3], "/")
								if len(detectArray) == 2 {
									replyString = replyString + "\n"
									var detectString string
									if dice > number {
										detectString = detectArray[1]
									} else {
										detectString = detectArray[0]
									}
									max := false
									if dice > 95 {
										max = true
									}
									diceResultString,diceResultInt := parseDiceArray(detectString,max)
									replyString = replyString + diceResultString + "\n《目前san值》" + strconv.Itoa(number) + "→" + strconv.Itoa(number - diceResultInt)
									if diceResultInt >= 5 {
										replyString = replyString + "\nsan值一次扣5以上 請骰靈感決定有沒有陷入暫時瘋狂"
									}
								}
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

func parseDiceArray(diceArrayString string, max bool) (replyString string,sum int){
	replyString = diceArrayString + "→"
	
	diceArrayString = strings.Replace(diceArrayString, "-", "+-",-1)
	
	diceArray := strings.Split(diceArrayString, "+")
	multiFlag := len(diceArray) > 1
	sum = 0
	for index, dice := range diceArray {

		diceLower := strings.ToLower(dice)
		diceForamt := strings.Split(diceLower, "d")
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
			plusFlag := true
			if diceNumber < 0 {
				plusFlag = false
				replyString = replyString + "-"
			}	
			if diceNumber > 1 || diceNumber < -1 {
				multiFlag = true
				replyString = replyString + "("
			}
			
			for i := 0.0 ; i < math.Abs(float64(diceNumber)) ; i++ {
				diceEachResult := diceType
				if !max {
					diceEachResult = rand.Intn(diceType) + 1
				}
				//GM
				loadData,loadErr := loadText()
				if loadErr == nil && loadData != ""{
					diceGM, GMErr := strconv.Atoi(loadData)
					if GMErr == nil && diceGM <= diceType{
						diceEachResult = diceGM
					}
					saveText("",true)
				}
				//GMend
				
				if plusFlag {
					sum = sum + diceEachResult
				}else{
					sum = sum - diceEachResult
				}
				if max {
					replyString = replyString + "大失敗取最大值"
				}
				replyString = replyString + strconv.Itoa(diceEachResult)
				replyString = replyString + "+"
			}
			if diceNumber > 1 || diceNumber < -1 {
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
	
	for ; strings.Contains(replyString, "++"); {
		replyString = strings.Replace(replyString, "++", "+",-1)
	}
	replyString = strings.Replace(replyString, "+)", ")",-1)
	replyString = strings.Replace(replyString, "+→", "→",-1)
	replyString = strings.Replace(replyString, "+-", "-",-1)
	replyString = strings.TrimSuffix(replyString, "+")
	return 
}
