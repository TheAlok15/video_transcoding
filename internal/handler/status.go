package handler

import (
	"net/http"

	"github.com/TheAlok15/video_transcoding/internal/database"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"github.com/gin-gonic/gin"
)

func GetJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	var job model.Job
	if err := database.DB.First(&job, "id = ?", jobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "job not found",
		})
	return
	}

	response := gin.H{
		"job_id": job.ID, "status": job.Status,
	}

	// only send url if completed
	if job.Status == "completed" {
		response["outputs"] = gin.H{
			"360p": job.Output360URL,
			"480p": job.Output480URL,
			"720p": job.Output720URL,
		}
	}

	if job.Status == "failed" {
		response["error"] = job.ErrorMessage
	}

	c.JSON(http.StatusOK, response)
}