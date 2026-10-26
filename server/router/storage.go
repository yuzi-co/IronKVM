package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/storage"
)

func storageRouter(r *gin.Engine) {
	service := storage.NewService()
	api := r.Group("/api").Use(middleware.CheckToken(), middleware.RequireRole(authn.RoleAdmin))

	api.GET("/storage/image", service.GetImages)                // get image list
	api.GET("/storage/image/mounted", service.GetMountedImage)  // get mounted image
	api.POST("/storage/image/mount", service.MountImage)        // mount image
	api.GET("/storage/cdrom", service.GetCdRom)                 // get CD-ROM flag
	api.GET("/storage/drives", service.GetDrives)               // list the virtual drives
	api.POST("/storage/drives/:id/insert", service.InsertDrive) // insert an image into a drive
	api.POST("/storage/drives/:id/eject", service.EjectDrive)   // eject a drive's image
	api.POST("/storage/image/delete", service.DeleteImage)      // delete image

	// The files inside an image: browse it, and exchange files with the host
	// through it while no drive serves it.
	api.GET("/storage/files", service.ListFiles)             // list a directory in an image
	api.GET("/storage/files/download", service.DownloadFile) // download a file from an image
	api.POST("/storage/files/upload", service.UploadFile)    // write a file into an image
	api.POST("/storage/files/delete", service.DeleteFile)    // delete a file or directory in an image
	api.GET("/storage/files/status", service.FileJobStatus)  // progress of the upload running
}
