package routes_jobs

import (
	"aura/jobs"
	"aura/logging"
	"aura/utils/httpx"
	"net/http"
)

func GetAllJobs(w http.ResponseWriter, r *http.Request) {
	ctx, ld := logging.CreateLoggingContext(r.Context(), r.URL.Path)
	logAction := ld.AddAction("Get All Scheduled Jobs", logging.LevelInfo)
	ctx = logging.WithCurrentAction(ctx, logAction)

	jobs := jobs.GetListOfJobs()
	httpx.SendResponse(w, ld, jobs)
}
