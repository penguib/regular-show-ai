package endpoints

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regular-show-ai/models"
	"regular-show-ai/util"

	"github.com/go-chi/render"
)

func GETScene(w http.ResponseWriter, r *http.Request) {

	if util.ScenesMetadata.CleanScenes == 0 {
		util.ScenesMetadata.CleanAllScenes()
	}

	var scene models.Scene
	var key string

	// need to loop since we are discarding the scenes that we
	// previously picked. so there is a chance we pick a random key
	// that was discarded
	for {
		random := rand.Intn(util.ScenesMetadata.SceneCount) + 1
		k := fmt.Sprint(random)
		s := util.ScenesMetadata.UnusedScenes[k]
		if !s.Dirty {
			scene = s
			key = k
			break
		}
	}

	render.JSON(w, r, scene)

	util.ScenesMetadata.DiscardScene(key)
}

type sceneRequest struct {
	Conversation []models.Speech
}

func (s *sceneRequest) Bind(r *http.Request) error {
	if len(s.Conversation) == 0 {
		return errors.New("Cannot post empty scene")
	}
	return nil
}

/*
func POSTScene(w http.ResponseWriter, r *http.Request) {
	data := &sceneRequest{}

	if err := render.Bind(r, data); err != nil {
		render.Status(r, http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("./scenes/%d", util.ScenesMetadata.SceneCount)
	os.Mkdir(path, 0755)

	render.Status(r, http.StatusOK)
	return
}
*/
