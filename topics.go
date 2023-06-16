package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"regular-show-ai/models"
	"regular-show-ai/util"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

const (
	uberduckAPI         = "https://api.uberduck.ai/speak"
	uberduckStatusAPI   = "https://api.uberduck.ai/speak-status"
	maxAudioWaitSeconds = 10
)

func readTopicsList() (string, error) {
	// kinda stupid we keep opening the file but whatever
	topics, err := os.OpenFile("./ideas.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return "", err
	}

	defer topics.Close()

	details, err := topics.Stat()
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(make([]byte, 0, details.Size()))

	_, err = topics.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(buf, topics)
	if err != nil {
		return "", err
	}

	// ReadBytes will return the first line, but now the buffer will
	// have the entire file contents except for the first line. Now
	// we should just copy these bytes into the file bytes and save it.
	topic, err := buf.ReadBytes('\n')

	if err != nil && err != io.EOF {
		return "", err
	}

	_, err = topics.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	nw, err := io.Copy(topics, buf)
	if err != nil {
		return "", err
	}

	err = topics.Truncate(nw)
	if err != nil {
		return "", err
	}

	err = topics.Sync()
	if err != nil {
		return "", err
	}

	return string(topic), nil
}

// currently only mordecai, rigby, and benson because
// their voices sound the best. Muscle man should come
// soon once the model is done
func randomCharacters() string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(2)
	if i == 0 {
		return "mordecai and rigby"
	}
	return "mordecai, rigby, and benson"
}

func generateConversation() ([]string, error) {
	topic, err := readTopicsList()
	if err != nil {
		panic(err)
	}

	// fallback if there was an error or there were no more
	// topics in ideas.txt
	if topic == "" {
		topic = "something funny"
	}

	prompt := fmt.Sprintf("generate a discussion between %s from regular show that is between 4 and 9 lines long that is about %s", randomCharacters(), topic)

	resp, err := chatGPT.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
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
	path := fmt.Sprintf("%s/%d", util.DiskPath, util.ScenesMetadata.SceneCount+1)
	fmt.Println(path)
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

func ffmpegWAVtoOGG(path string) {
	newWAVPath := strings.Replace(path, "temp-", "", -1)
	newWAVPath = strings.Replace(newWAVPath, "wav", "ogg", -1)
	err := ffmpeg_go.Input(path, nil).Output(newWAVPath, nil).Run()
	if err != nil {
		panic(err)
	}
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

			audioPath := fmt.Sprintf("%s/audio/temp-%d.wav", dpath, i)
			file, err := os.Create(audioPath)
			if err != nil {
				panic(err)
			}

			_, err = io.Copy(file, res.Body)
			if err != nil {
				panic(err)
			}

			ffmpegWAVtoOGG(audioPath)

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

	// generateAudio(scene.Conversation, path)
}

func GenerateTopics() {
	for {
		generateScenes()
		time.Sleep(1 * time.Minute)
	}
}
