package storage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ego-component/egorm"
	"github.com/spf13/cast"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/pkg/agent/search"
	"github.com/clickvisual/clickvisual/api/internal/pkg/component/core"
	"github.com/clickvisual/clickvisual/api/internal/pkg/constx"
	"github.com/clickvisual/clickvisual/api/internal/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/internal/pkg/model/view"
	"github.com/clickvisual/clickvisual/api/internal/pkg/utils/mapping"
	"github.com/clickvisual/clickvisual/api/internal/service"
	"github.com/clickvisual/clickvisual/api/internal/service/event"
	"github.com/clickvisual/clickvisual/api/internal/service/permission"
	"github.com/clickvisual/clickvisual/api/internal/service/permission/pmsplugin"
)

// KafkaJsonMapping  godoc
// @Summary	     Kafka JSON field mapping
// @Description  Kafka JSON field mapping
// @Tags         LOGSTORE
// @Accept       json
// @Produce      json
// @Param        req query view.ReqKafkaJSONMapping true "params"
// @Success      200 {object} view.List
// @Router       /api/v2/storage/mapping-json [post]
func KafkaJsonMapping(c *core.Context) {
	var req view.ReqKafkaJSONMapping
	if err := c.Bind(&req); err != nil {
		c.JSONE(1, "request parameter error: "+err.Error(), nil)
		return
	}
	res, err := mapping.Handle(req.Data, false)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), nil)
		return
	}
	c.JSONOK(res)
}

// Create  godoc
// @Summary	     Creating a log library
// @Description  Creating a log library
// @Tags         LOGSTORE
// @Accept       json
// @Produce      json
// @Param        req query view.ReqStorageCreate true "params"
// @Success      200 {object} core.Res{}
// @Router       /api/v2/storage [post]
func Create(c *core.Context) {
	var param view.ReqStorageCreate
	err := c.Bind(&param)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), nil)
		return
	}
	databaseInfo, err := db.DatabaseInfo(invoker.Db, param.DatabaseId)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), nil)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(databaseInfo.Iid),
		SubResource: pmsplugin.Log,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixDatabase,
		DomainId:    strconv.Itoa(databaseInfo.ID),
	}); err != nil {
		c.JSONE(1, "permission verification failed", err)
		return
	}
	param.CreateType = constx.TableCreateTypeJSONEachRow
	_, err = service.StorageCreate(c.Uid(), databaseInfo, param)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), err)
		return
	}
	event.Event.InquiryCMDB(c.User(), db.OpnTablesCreate, map[string]interface{}{"param": param})
	c.JSONOK()
}

// AnalysisFields  godoc
// @Summary	     iStorage analysis field list
// @Description  iStorage analysis field list
// @Tags         LOGSTORE
// @Accept       json
// @Produce      json
// @Param        storage-id path int true "table id"
// @Success      200 {object} view.RespStorageAnalysisFields
// @Router       /api/v2/storage/{storage-id}/analysis-fields [get]
func AnalysisFields(c *core.Context) {
	storageId := cast.ToInt(c.Param("storage-id"))
	if storageId == 0 {
		c.JSONE(1, "invalid parameter", nil)
		return
	}
	res := view.RespStorageAnalysisFields{
		BaseFields: make([]view.StorageAnalysisField, 0),
		LogFields:  make([]view.StorageAnalysisField, 0),
	}
	// Read the index data
	conds := egorm.Conds{}
	conds["tid"] = storageId
	fields, _ := db.IndexList(conds)
	for _, row := range fields {
		f := view.StorageAnalysisField{
			Id:         row.ID,
			Tid:        row.Tid,
			Field:      row.Field,
			RootName:   row.RootName,
			Typ:        row.Typ,
			HashTyp:    row.HashTyp,
			Alias:      row.Alias,
			Ctime:      row.Ctime,
			Utime:      row.Utime,
			OrderField: fmt.Sprintf("%s.%s", row.RootName, row.Field),
		}
		if row.Kind == 0 {
			res.BaseFields = append(res.BaseFields, f)
		} else {
			res.LogFields = append(res.LogFields, f)
		}
	}
	// keys sort by the first letter
	sort.Slice(res.BaseFields, func(i, j int) bool {
		return res.BaseFields[i].OrderField < res.BaseFields[j].OrderField
	})
	sort.Slice(res.LogFields, func(i, j int) bool {
		return res.LogFields[i].OrderField < res.LogFields[j].OrderField
	})
	c.JSONOK(res)
}

// Update  godoc
// @Summary	     iStorage update
// @Description  iStorage update
// @Tags         LOGSTORE
// @Accept       json
// @Produce      json
// @Param        storage-id path int true "table id"
// @Param        req query view.ReqStorageUpdate true "params"
// @Success      200 {object} core.Res{}
// @Router       /api/v2/storage/{storage-id} [patch]
func Update(c *core.Context) {
	id := cast.ToInt(c.Param("storage-id"))
	if id == 0 {
		c.JSONE(1, "invalid parameter", nil)
		return
	}
	var (
		req view.ReqStorageUpdate
		err error
	)
	if err = c.Bind(&req); err != nil {
		c.JSONE(1, "invalid parameter: "+err.Error(), nil)
		return
	}
	tableInfo, err := db.TableInfo(invoker.Db, id)
	if err != nil {
		c.JSONE(1, "invalid parameter: "+err.Error(), nil)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(tableInfo.Database.Iid),
		SubResource: pmsplugin.Log,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixTable,
		DomainId:    strconv.Itoa(id),
	}); err != nil {
		c.JSONE(1, "permission verification failed", err)
		return
	}
	op, err := service.InstanceManager.Load(tableInfo.Database.Iid)
	if err != nil {
		c.JSONE(1, "update failed 01: "+err.Error(), nil)
		return
	}
	// check merge tree
	if req.MergeTreeTTL != tableInfo.Days {
		// alert merge tree engine table
		if err = op.UpdateMergeTreeTable(&tableInfo, req); err != nil {
			c.JSONE(1, "update failed 02: "+err.Error(), nil)
			return
		}
	}
	var streamSQL string
	// check kafka
	if req.KafkaSkipBrokenMessages != tableInfo.KafkaSkipBrokenMessages ||
		req.KafkaBrokers != tableInfo.Brokers ||
		req.KafkaConsumerNum != tableInfo.ConsumerNum ||
		req.KafkaTopic != tableInfo.Topic {
		// drop & create kafka engine table
		if streamSQL, err = op.CreateKafkaTable(&tableInfo, req); err != nil {
			c.JSONE(1, "update failed 03: "+err.Error(), nil)
			return
		}

	}
	// just mysql record update
	ups := make(map[string]interface{}, 0)
	ups["uid"] = c.Uid()
	ups["days"] = req.MergeTreeTTL
	ups["topic"] = req.KafkaTopic
	ups["brokers"] = req.KafkaBrokers
	ups["consumer_num"] = req.KafkaConsumerNum
	ups["desc"] = req.Desc
	ups["kafka_skip_broken_messages"] = req.KafkaSkipBrokenMessages
	ups["v3_table_type"] = req.V3TableType
	if streamSQL != "" {
		ups["sql_stream"] = streamSQL
	}
	if tableInfo.V3TableType != req.V3TableType {
		if req.V3TableType == db.V3TableTypeJaegerJSON {
			err = op.CreateTraceJaegerDependencies(tableInfo.Database.Name, tableInfo.Database.Cluster, tableInfo.Name, tableInfo.Days)
			if err != nil {
				c.JSONE(1, "update failed 04: "+err.Error(), nil)
				return
			}
		} else {
			err = op.DeleteTraceJaegerDependencies(tableInfo.Database.Name, tableInfo.Database.Cluster, tableInfo.Name)
			if err != nil {
				c.JSONE(1, "update failed 05: "+err.Error(), nil)
				return
			}
		}
	}
	// 判断是否增加依赖解析
	if err = db.TableUpdate(invoker.Db, id, ups); err != nil {
		c.JSONE(1, "update failed 06: "+err.Error(), nil)
		return
	}
	event.Event.InquiryCMDB(c.User(), db.OpnTablesUpdate, map[string]interface{}{"req": req})
	c.JSONOK()
}

// CreateStorageByTemplate  godoc
// @Summary	     Create storage by template
// @Description  Create storage by template
// @Tags         LOGSTORE
// @Accept       json
// @Produce      json
// @Param        template path string true "template"
// @Param        req query view.ReqCreateStorageByTemplateEgo true "params"
// @Success      200 {object} core.Res{}
// @Router       /api/v2/storage/{template} [post]
func CreateStorageByTemplate(c *core.Context) {
	tpl := strings.TrimSpace(c.Param("template"))
	switch tpl {
	case "ego":
		createStorageByTemplateEgo(c)
		return
	case "ilogtail":
		createStorageByTemplateILogtail(c)
		return
	case "agent":
		createStorageByTemplateAgent(c)
		return
	}
	c.JSONE(core.CodeErr, "template error", nil)
}

func createStorageByTemplateEgo(c *core.Context) {
	var param view.ReqCreateStorageByTemplateEgo
	err := c.Bind(&param)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), err)
		return
	}
	databaseInfo, err := db.DatabaseInfo(invoker.Db, param.DatabaseId)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), err)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(databaseInfo.Iid),
		SubResource: pmsplugin.Log,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixDatabase,
		DomainId:    strconv.Itoa(databaseInfo.ID),
	}); err != nil {
		c.JSONE(1, "permission verification failed", err)
		return
	}
	if err = service.Storage.CreateByEgoTemplate(c.Uid(), databaseInfo, param); err != nil {
		c.JSONE(core.CodeErr, err.Error(), err)
		return
	}
	event.Event.InquiryCMDB(c.User(), db.OpnTablesCreate, map[string]interface{}{"param": param})
	c.JSONOK()
}

func createStorageByTemplateAgent(c *core.Context) {
	var param view.ReqCreateAgentStorage
	err := c.Bind(&param)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), err)
		return
	}
	databaseInfo, err := db.DatabaseInfo(invoker.Db, param.DatabaseId)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), err)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(databaseInfo.Iid),
		SubResource: pmsplugin.Log,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixDatabase,
		DomainId:    strconv.Itoa(databaseInfo.ID),
	}); err != nil {
		c.JSONE(1, "permission verification failed", err)
		return
	}
	tx := invoker.Db.Begin()
	conds := egorm.Conds{}
	conds["did"] = databaseInfo.ID
	conds["name"] = param.Name
	tableInfo, _ := db.TableInfoX(tx, conds)
	if tableInfo.ID != 0 {
		tx.Rollback()
		c.JSONE(1, "table is repeat", err)
		return
	}
	tableInfo = db.BaseTable{
		Did:  databaseInfo.ID,
		Name: param.Name,
	}
	err = db.TableCreate(tx, &tableInfo)
	if err != nil {
		tx.Rollback()
		c.JSONE(1, "table created failed", err)
		return
	}
	for _, col := range search.DefaultBaseFields {
		err = db.IndexCreate(tx, &db.BaseIndex{
			Tid:      tableInfo.ID,
			Field:    col.Name,
			Typ:      col.Type,
			Alias:    "",
			RootName: "",
			Kind:     0,
		})
		if err != nil {
			tx.Rollback()
			c.JSONE(1, "index created failed", err)
			return
		}
	}
	for _, col := range search.DefaultLogFields {
		err = db.IndexCreate(tx, &db.BaseIndex{
			Tid:      tableInfo.ID,
			Field:    col.Name,
			Typ:      col.Type,
			Alias:    "",
			RootName: "",
			Kind:     1,
		})
		if err != nil {
			tx.Rollback()
			c.JSONE(1, "index created failed", err)
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		c.JSONE(core.CodeErr, "commit error", err)
		return
	}
	event.Event.InquiryCMDB(c.User(), db.OpnTablesCreate, map[string]interface{}{"param": param})
	c.JSONOK()
}

func createStorageByTemplateILogtail(c *core.Context) {
	var param view.ReqCreateStorageByTemplateILogtail
	err := c.Bind(&param)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), err)
		return
	}
	databaseInfo, err := db.DatabaseInfo(invoker.Db, param.DatabaseId)
	if err != nil {
		c.JSONE(core.CodeErr, "invalid parameter: "+err.Error(), err)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(databaseInfo.Iid),
		SubResource: pmsplugin.Log,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixDatabase,
		DomainId:    strconv.Itoa(databaseInfo.ID),
	}); err != nil {
		c.JSONE(1, "permission verification failed", err)
		return
	}
	if err = service.Storage.CreateByILogtailTemplate(c.Uid(), databaseInfo, param); err != nil {
		c.JSONE(core.CodeErr, err.Error(), err)
		return
	}
	event.Event.InquiryCMDB(c.User(), db.OpnTablesCreate, map[string]interface{}{"param": param})
	c.JSONOK()
}
