package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regular-show-ai/models"
)

var DiskPath string

type Scenes struct {
	CleanScenes  int
	SceneCount   int
	UnusedScenes map[string]models.Scene
}

func (s *Scenes) Init(dpath string, generate bool) error {
	if dpath == "" {
		DiskPath = "./scenes"
	} else {
		DiskPath = dpath
	}

	files, err := ioutil.ReadDir(dpath)

	if err != nil {
		return errors.New(fmt.Sprintf("Problem with trying reading scenes directory: %s", DiskPath))
	}

	if len(files) == 0 && !generate {
		return errors.New("No scenes")
	}

	s.UnusedScenes = make(map[string]models.Scene)
	fileCount := 0

	// If there are no scenes, we should still be able to generate them if the
	// generate flag was provided
	if len(files) == 0 && generate {
		s.CleanScenes = 0
		s.SceneCount = 0

		return nil
	}

	for _, v := range files {
		if !v.IsDir() {
			continue
		}

		path := fmt.Sprintf("%s/%s/metadata.json", DiskPath, v.Name())
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
