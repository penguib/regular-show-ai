package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"regular-show-ai/models"
	"regular-show-ai/util"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	uberduckAPI         = "https://api.uberduck.ai/speak"
	uberduckStatusAPI   = "https://api.uberduck.ai/speak-status"
	maxAudioWaitSeconds = 10
)

func generateConversation() ([]string, error) {
	resp, err := chatGPT.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "create a funny discussion between mordecai and rigby from regular show that is at most 4 lines long",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	splitContent := strings.Split(content, "\n")

	return splitContent, nil
}

func setUpDir() string {
	path := fmt.Sprintf("%s/scenes/%d", util.DiskPath, util.ScenesMetadata.SceneCount+1)
	os.Mkdir(path, 0755)

	audioPath := fmt.Sprintf("%s/audio", path)
	os.Mkdir(audioPath, 0755)

	return path
}

func checkAudioStatus(audioUUID string) (bool, string) {
	for i := 0; i < maxAudioWaitSeconds; i++ {
		req, err := http.NewRequest("GET", uberduckStatusAPI, nil)
		if err != nil {
			panic(err)
		}

		q := req.URL.Query()
		q.Add("uuid", audioUUID)
		req.URL.RawQuery = q.Encode()

		req.SetBasicAuth(os.Getenv("UBERDUCK_PUBLIC_KEY"), os.Getenv("UBERDUCK_PRIVATE_KEY"))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return false, ""
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var jsonResp map[string]interface{}
		json.Unmarshal(body, &jsonResp)

		if jsonResp["path"] != nil {
			return true, jsonResp["path"].(string)
		}

		time.Sleep(1 * time.Second)
	}
	return false, ""
}

func generateAudio(speech []models.Speech, dpath string) {
	for i, v := range speech {
		voiceUUID := models.GetCharacterVoiceUUID(v.Character)
		payload, _ := json.Marshal(map[string]interface{}{
			"speech":          v.Content,
			"voicemodel_uuid": voiceUUID,
		})
		req, err := http.NewRequest("POST", uberduckAPI, bytes.NewBuffer(payload))
		if err != nil {
			panic(err)
		}
		req.SetBasicAuth(os.Getenv("UBERDUCK_PUBLIC_KEY"), os.Getenv("UBERDUCK_PRIVATE_KEY"))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		var jsonResp map[string]interface{}
		json.Unmarshal(body, &jsonResp)
		audioUUID := jsonResp["uuid"]

		status, path := checkAudioStatus(audioUUID.(string))
		if status {
			res, err := http.Get(path)
			if err != nil {
				panic(err)
			}

			defer res.Body.Close()

			file, err := os.Create(fmt.Sprintf("%s/audio/%d.wav", dpath, i))
			if err != nil {
				panic(err)
			}

			_, err = io.Copy(file, res.Body)
			if err != nil {
				panic(err)
			}
		}
	}
}

func generateScenes() {
	content, err := generateConversation()
	if err != nil {
		return
	}

	r, err := regexp.Compile("([a-zA-Z]+):")
	if err != nil {
		return
	}

	var scene models.Scene
	characters := 0

	for _, v := range content {
		match := r.FindString(v)

		// sometimes theres a literal empty string for some reason
		if len(match) == 0 || match == "" {
			continue
		}

		character := strings.Trim(match[:len(match)-1], " ")
		characters |= int(models.GetCharacterByName(character))

		speech := strings.Trim(v[len(match):], " ")

		scene.Conversation = append(scene.Conversation, models.Speech{
			Character: models.GetCharacterByName(character),
			Content:   speech,
		})
	}

	scene.ID = util.ScenesMetadata.SceneCount + 1
	scene.Characters = characters

	path := setUpDir()
	file, _ := json.MarshalIndent(scene, "", "	")
	err = ioutil.WriteFile(fmt.Sprintf("%s/metadata.json", path), file, 0644)
	if err != nil {
		panic(err)
	}

	scene.Dirty = false
	util.ScenesMetadata.UnusedScenes[fmt.Sprint(scene.ID)] = scene
	util.ScenesMetadata.SceneCount++

	generateAudio(scene.Conversation, path)
}

func GenerateTopics() {
	for {
		generateScenes()
		time.Sleep(1 * time.Minute)
	}
}
