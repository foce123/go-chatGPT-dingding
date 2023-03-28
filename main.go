package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
)

type Msg struct {
	SenderID string            `json:"senderStaffId"`
	Text     map[string]string `json:"text"`
}
type DData struct {
	At   map[string][]string `json:"at"`
	Text map[string]string   `json:"text"`
	//Markdown map[string]string   `json:"markdown"`
	Msgtype string `json:"msgtype"`
}
type Respchat struct {
	Choices []map[string]interface{} `json:"choices"`
}
type Chatdata struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float32             `json:"temperature"`
}

// keypoint
func JSONDecode(r io.Reader, obj interface{}) error {
	if err := json.NewDecoder(r).Decode(obj); err != nil {
		return err
	}
	return nil
}

func Handler(c *gin.Context) {
	Apikey := os.Args[2]
	DDToken := os.Args[3]
	if strings.HasPrefix(Apikey, `sk-`) && len(DDToken) != 0 {
		fmt.Println("key and token is right")
	} else {
		fmt.Println("key and token is invalid")
		os.Exit(1)
	}
	// print body
	var json Msg
	data, err := c.GetRawData()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("data:", string(data))
	// reput to body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	err2 := JSONDecode(c.Request.Body, &json)
	if err2 != nil {
		fmt.Println("decode err:", err2)
	}
	// print params
	fmt.Println("userid:", json.SenderID, "text:", json.Text["content"])
	c.JSON(http.StatusOK, "success")
	chatdata := ReqChatGPT(Apikey, json.Text["content"])
	ToDingding(DDToken, json.SenderID, chatdata)
}

func ReqChatGPT(apikey string, message string) string {
	client := &http.Client{}
	var reqchat Chatdata
	reqchat.Model = "gpt-3.5-turbo"
	reqchat.Temperature = 0.5
	reqchat.Messages = []map[string]string{
		{"role": "user", "content": message},
	}
	chaturl := "https://api.openai.com/v1/chat/completions"

	reqData, _ := json.Marshal(reqchat)
	req, _ := http.NewRequest("POST", chaturl, bytes.NewBuffer(reqData))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apikey)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	var jchat Respchat
	sbody := io.NopCloser(bytes.NewBuffer(body))
	err2 := JSONDecode(sbody, &jchat)
	if err2 != nil {
		fmt.Println("jchat decode err:", err2)
	}
	msgdata := jchat.Choices[0]["message"].(map[string]interface{})
	fmt.Println(msgdata["content"].(string))
	return msgdata["content"].(string)
}

func ToDingding(ddtoken string, userid string, data string) {
	dd_url := "https://oapi.dingtalk.com/robot/send?access_token="
	var pdata DData
	pdata.At = map[string][]string{
		"atUserIds": {userid},
	}
	pdata.Text = map[string]string{
		"content": data,
	}
	pdata.Msgtype = "text"
	sdata, err := json.Marshal(pdata)
	if err != nil {
		fmt.Printf("Map to Byte_array, exception:%s\n", err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", dd_url+ddtoken, bytes.NewBuffer(sdata))
	req.Header.Add("Content-Type", "application/json")
	resp, err1 := client.Do(req)
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("post to dingding error, exception:%s\n", err)
	}
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/", Handler)
	r.Run(":" + os.Args[1])
}
