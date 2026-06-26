package main

import (
	"embed"
	"log"

	"github.com/k3desktop/k3desktop/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "K3Desktop",
		Description: "k3d cluster manager",
		Services: []application.Service{
			application.NewService(&service.ClusterService{}),
			application.NewService(&service.NodeService{}),
			application.NewService(&service.RegistryService{}),
			application.NewService(&service.KubeconfigService{}),
			application.NewService(&service.VersionService{
				AppVersion:    Version,
				AppBuildDate:  BuildDate,
				AppCommitHash: CommitHash,
			}),
			application.NewService(&service.ProfileService{}),
			application.NewService(&service.LogService{}),
			application.NewService(&service.BlueprintService{}),
			application.NewService(&service.OperationsService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Linux: application.LinuxOptions{
			ProgramName: "k3desktop",
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "K3Desktop",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
