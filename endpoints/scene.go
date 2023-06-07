package endpoints

import (
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
