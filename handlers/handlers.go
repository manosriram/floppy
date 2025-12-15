package handlers

import (
	"github.com/manosriram/floppy/handlers/api"
	"github.com/manosriram/floppy/internal/config"
)

type ApiHandler struct {
	ApiFileHandler       api.FiberApiFileHandler
	ApiMountpointHandler api.FiberApiMountpointHandler
}

func NewApiHandler(m config.Mountpoints) ApiHandler {
	apiFileHandler := api.FiberApiFileHandler{}
	apiMountpointHandler := api.FiberApiMountpointHandler{
		M: m,
	}

	return ApiHandler{
		ApiFileHandler:       apiFileHandler,
		ApiMountpointHandler: apiMountpointHandler,
	}
}
