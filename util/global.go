package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regular-show-ai/models"
)

var diskPath string

type Scenes struct {
	CleanScenes  int
	SceneCount   int
	UnusedScenes map[string]models.Scene
}

func (s *Scenes) Init(dpath string) error {
	if dpath == "" {
		diskPath = "./scenes"
	} else {
		diskPath = dpath
	}

	files, err := ioutil.ReadDir(dpath)

	if err != nil {
		return errors.New(fmt.Sprintf("Problem with trying reading scenes directory: %s", diskPath))
	}

	if len(files) == 0 {
		return errors.New("No scenes")
	}

	s.UnusedScenes = make(map[string]models.Scene)
	fileCount := 0

	for _, v := range files {
		if !v.IsDir() {
			continue
		}

		path := fmt.Sprintf("%s/%s/metadata.json", diskPath, v.Name())
		metadata, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		defer metadata.Close()

		bytes, err := ioutil.ReadAll(metadata)
		if err != nil {
			panic(err)
		}

		var scene models.Scene
		err = json.Unmarshal(bytes, &scene)
		if err != nil {
			panic(err)
		}

		scene.Dirty = false

		s.UnusedScenes[v.Name()] = scene
		fileCount++
	}

	s.CleanScenes = fileCount
	s.SceneCount = fileCount

	return nil
}

func (s *Scenes) DiscardScene(id string) {
	if entry, ok := s.UnusedScenes[id]; ok {
		entry.Dirty = true
		s.UnusedScenes[id] = entry
	}
	s.CleanScenes--
}

func (s *Scenes) CleanAllScenes() {
	for _, v := range s.UnusedScenes {
		v.Dirty = false
		s.UnusedScenes[fmt.Sprint(v.ID)] = v
	}
	s.CleanScenes = s.SceneCount
}

var ScenesMetadata Scenes
