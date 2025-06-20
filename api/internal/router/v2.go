package router

import (
	_ "embed"

	"github.com/gin-gonic/gin"

	goredoc "github.com/link-duan/go-redoc"

	"github.com/clickvisual/clickvisual/api/docs"
	"github.com/clickvisual/clickvisual/api/internal/api/apiv2/alert"
	"github.com/clickvisual/clickvisual/api/internal/api/apiv2/base"
	"github.com/clickvisual/clickvisual/api/internal/api/apiv2/pandas"
	"github.com/clickvisual/clickvisual/api/internal/api/apiv2/storage"
	"github.com/clickvisual/clickvisual/api/internal/pkg/component/core"
	"github.com/clickvisual/clickvisual/api/internal/router/middlewares"
)

// Defines interface prefixes in terms of module overrides：
// The global basic readable information module - base
// The log module - storage
// The alert module - alert
// The data analysis module - pandas
// The configuration module - cmdb
// The system management module - sysop
func v2(r *gin.RouterGroup) {
	r = r.Group("/api/v2", middlewares.AuthChecker())
	// swagger docs
	{
		r.GET("/swagger/*any", goredoc.GinHandler(&goredoc.Setting{
			OpenAPIJson: docs.EGOGenAPI,
			UriPrefix:   "/api/v2/swagger",
			Title:       "Go ReDoc",
			RedocOptions: map[string]string{
				"schema-expansion-level": "all",
				"expand-responses":       "200,201",
			},
		}))
	}
	// The global basic readable information module - base
	{
		// user apis
		r.GET("/base/users", core.Handle(base.ListUser))
		r.POST("/base/users", core.Handle(base.CreateUser))
		r.PATCH("/base/users/:user-id", core.Handle(base.UpdateUser))
		r.DELETE("/base/users/:user-id", core.Handle(base.DeleteUser))
		r.PATCH("/base/users/:user-id/password-reset", core.Handle(base.ResetUserPassword))
		// other apis
		r.GET("/base/instances", core.Handle(base.InstanceList))
		// todo: deprecated
		r.POST("/base/shorturls", core.Handle(base.ShortURLCreate))
		r.GET("/base/su/:s-code", core.Handle(base.ShortURLRedirect))
		// instance
		r.GET("/base/install/local", core.Handle(base.ListUser))
	}
	// The data analysis module - pandas
	{
		// The edit lock can be actively obtained if the file is in the edit state
		r.POST("/pandas/nodes/:node-id/lock-acquire", core.Handle(pandas.NodeLockAcquire))
		// Scheduled Task Scheduling
		r.POST("/pandas/nodes/:node-id/crontab", core.Handle(pandas.NodeCrontabCreate))
		r.PATCH("/pandas/nodes/:node-id/crontab", core.Handle(pandas.NodeCrontabUpdate))
		// The node running data is processed by Excel
		r.GET("/pandas/nodes/:node-id/results", core.Handle(pandas.NodeResultListPage))
		r.PATCH("/pandas/nodes-results/:result-id", core.Handle(pandas.NodeResultUpdate))
		// Timing schedule stats
		r.GET("/pandas/workers", core.Handle(pandas.WorkerList))
		r.GET("/pandas/workers/dashboard", core.Handle(pandas.WorkerDashboard))
		r.GET("/pandas/instances/:instance-id/table-dependencies", core.Handle(pandas.TableDependencies))
		// DDL structural transfer
		r.POST("/pandas/utils/structural-transfer", core.Handle(pandas.StructuralTransfer))
		// TableName Create SQL
		r.GET("/pandas/instances/:instance-id/databases/:database/tables/:table/create-sql", core.Handle(pandas.TableCreateSQL))
	}
	// The log module - storage
	{
		r.POST("/storage", core.Handle(storage.Create))
		r.PATCH("/storage/:storage-id", core.Handle(storage.Update))
		r.POST("/storage/mapping-json", core.Handle(storage.KafkaJsonMapping))
		r.POST("/storage/:template", core.Handle(storage.CreateStorageByTemplate))
		r.GET("/storage/:storage-id/analysis-fields", core.Handle(storage.AnalysisFields))
		// trace apis
		r.GET("/storage/traces", core.Handle(storage.GetTraceList))
		r.PATCH("/storage/:storage-id/trace", core.Handle(storage.UpdateTraceInfo))
		r.GET("/storage/:storage-id/trace-graph", core.Handle(storage.GetTraceGraph))
		r.GET("/storage/:storage-id/columns", core.Handle(storage.GetStorageColumns))
		// collect
		r.GET("/storage/collects", core.Handle(storage.ListCollect))
		r.POST("/storage/collects", core.Handle(storage.CreateCollect))
		r.PATCH("/storage/collects/:collect-id", core.Handle(storage.UpdateCollect))
		r.DELETE("/storage/collects/:collect-id", core.Handle(storage.DeleteCollect))
	}
	// The log module - alert
	{
		r.GET("/alert/settings", core.Handle(alert.SettingList))
		r.GET("/alert/settings/:instance-id", core.Handle(alert.SettingInfo))
		r.POST("/alert/metrics-samples", core.Handle(alert.CreateMetricsSamples))
		r.PATCH("/alert/settings/:instance-id", core.Handle(alert.SettingUpdate))
	}
}
