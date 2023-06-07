package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"regular-show-ai/models"
	"regular-show-ai/util"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
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

func SetUpDir() string {
	path := fmt.Sprintf("./scenes/%d", util.ScenesMetadata.SceneCount+1)
	os.Mkdir(path, 0755)

	audioPath := fmt.Sprintf("%s/audio", path)
	os.Mkdir(audioPath, 0755)

	return path
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

	path := SetUpDir()
	file, _ := json.MarshalIndent(scene, "", "	")
	err = ioutil.WriteFile(fmt.Sprintf("%s/metadata.json", path), file, 0644)
	if err != nil {
		panic(err)
	}

	scene.Dirty = false
	util.ScenesMetadata.UnusedScenes[fmt.Sprint(scene.ID)] = scene
	util.ScenesMetadata.SceneCount++
}

func loopScenes() {
}

func GenerateTopics() {
	for {
		generateScenes()
		time.Sleep(1 * time.Minute)
	}
}
