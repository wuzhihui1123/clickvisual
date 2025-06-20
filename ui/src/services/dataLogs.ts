import { TimeBaseType } from "@/services/systemSetting";
import { request } from "@umijs/max";
import { message } from "antd";
export interface QueryLogsProps {
  st: number;
  et: number;
  query?: string | undefined;
  pageSize?: number;
  page?: number;
  alarmMode?: number;
  filters?: string[];
}

export interface GetTableIdRequest {
  instance: string;
  database: string;
  datasource: string;
  table: string;
}

export interface StatisticalTableResponse {
  isNeedSort: boolean;
  logs: any[];
  sortRule: string[];
}

export interface LogsResponse extends StatisticalTableResponse {
  count: number;
  cost: number;
  hiddenFields: string[];
  defaultFields: string[];
  keys: IndexInfoType[];
  limited: number;
  terms: string[][];
  query: string;
  where: string;
  isTrace: number; // 0 非链路模式 1 jaeger_json格式
}

export interface ViewResponse {
  id: number;
  viewName: string;
}

export interface CreatedLogLibraryRequestEachRow {
  databaseId: number;
  tableName: string;
  typ: number;
  days: number;
  brokers: string;
  topics: string;
  consumers: number;
  kafkaSkipBrokenMessages?: number;
  desc?: string;
  timeField: string;
  rawLogField: string;
  source: string;
}

export interface CreatedLogLibraryRequestAsString {
  databaseId: number;
  tableName: string;
  timeFieldType?: number;
  days: number;
  brokers: string;
  topics: string;
  consumers: number;
  kafkaSkipBrokenMessages?: number;
  desc?: string;
  timeField?: string;
  isKafkaTimestamp: number; // 1 yes 0 no
  v3TableType: number; // 0 default 1 jaegerJson
}

export interface CreatedTableTemplateType {
  brokers: string;
  databaseId: number;
  topicsApp: string;
  topicsEgo: string;
  topicsIngressStderr: string;
  topicsIngressStdout: string;
}

export interface CreatedViewRequest {
  id?: number;
  viewName: string;
  isUseDefaultTime: number;
  key?: string;
  format?: string;
}

export interface ViewInfoResponse extends TimeBaseType {
  id: number;
  uid: number;
  tid: number;
  name: string;
  isUseDefaultTime: number;
  key: string;
  format: string;
  sql_view: string;
}

export interface HighChartsResponse {
  count: number;
  progress: string;
  histograms: HighCharts[];
}

export interface HighCharts {
  count: number;
  from: number;
  to: number;
  progress: string;
}

export interface DatabaseResponse {
  datasourceType: string;
  id: number;
  iid: number;
  instanceName: string;
  name: string;
  uid?: number;
  mode: number;
  cluster?: string[];
  desc: string;
  instanceDesc: string;
}

export interface TablesResponse {
  id: number;
  tableName: string;
  createType: number;
  desc: string;
  relTraceTableId: number;
}

export interface TableInfoResponse {
  brokers: string;
  createType: number;
  timeField: string;
  days: number;
  did: number;
  name: string;
  sqlContent: TableSqlContent;
  topic: string;
  typ: number;
  uid: number;
  desc: string;
  database: DatabaseResponse;
  traceTableId: number;
}

export interface TableSqlContent {
  keys: string[];
  data: any;
}

export interface LocalTables {
  name: string;
  tables: { name: string }[];
}

export interface TableColumnsRequest {
  databaseName: string;
  tableName: string;
}

export interface TableColumn {
  name: string;
  type: number;
  typeDesc: string;
}

export interface TableColumnsResponse {
  index: number;
  all: TableColumn[];
  conformToStandard: TableColumn[];
}

export interface CreateLocalTableRequest {
  databaseName: string;
  tableName: string;
  timeField: string;
  timeFieldType: number;
}

export interface CreateLocalTableRequestBatch {
  timeField: string;
  mode: number;
  instance: number;
  tableList: TableListType;
  cluster: string;
}

export interface TableListType {
  timeField: number;
  desc: string;
  timeFieldType?: number;
  databaseName: string;
  tableName: string;
}

export interface IndexInfoType {
  id?: number;
  tid?: number;
  field: string;
  alias: string;
  typ: number;
  rootName: string;
  jsonIndex: IndexInfoType[];
  hashTyp: number;
  ctime: number;
  utime: number;
}

export interface IndexRequest {
  data?: IndexInfoType[];
}

export interface IndexDetailRequest {
  st: number;
  et: number;
  query?: string | undefined;
}

export interface IndexDetail {
  count: number;
  indexName: string;
  percent: number;
}

export interface UpdateTableInfoType {
  desc?: string;
  kafkaBrokers?: string;
  kafkaConsumerNum?: 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8;
  kafkaSkipBrokenMessages?: string;
  kafkaTopic?: string;
  mergeTreeTTL?: number;
}

export interface LogFilterType {
  tableId: any;
  alias: string;
  ctime: number;
  dtime: number;
  id: number;
  statement: string;
  uid: number;
  utime: number;
  collectType: CollectType;
}

export enum CollectType {
  query = 1,
  tableFilter = 2,
  globalFilter = 4,
  allFilter = 6,
}

export default {
  // Get chart information
  async getHighCharts(
    tableId: number,
    params: QueryLogsProps,
    cancelToken: any
  ) {
    if (params.query && params.et && params.st) {
      let queryLower = params.query.toLowerCase()
      // 留个口子， 如果包含查询魔数，则不做查询时间段限制
      if (queryLower.indexOf("6=6 and") < 0) {
        let maxQueryRange = 3600 * 24 * 15 + 300;
        let errorMsg = "包含查询条件的情况下，查询时间段不允许超过15天。";
        if (queryLower.indexOf("like") > 0) {
          maxQueryRange = 3600 * 8 + 300;
          errorMsg = "查询条件中包含 like 的情况下，查询时间段不允许超过8小时。";
        }
        if(params.et - params.st > maxQueryRange) {
          message.error(errorMsg, 5);
          return;
        }
      }
    }
    return request<API.Res<HighChartsResponse>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tableId}/charts`,
      {
        cancelToken,
        method: "GET",
        params,
        skipErrorHandler: true,
      }
    );
  },

  // Get log information
  async getLogs(tableId: number, params: QueryLogsProps, cancelToken: any) {
    if (params.query && params.et && params.st) {
      let queryLower = params.query.toLowerCase()
      // 留个口子， 如果包含查询魔数，则不做查询时间段限制
      if (queryLower.indexOf("6=6 and") < 0) {
        let maxQueryRange = 3600 * 24 * 15 + 300;
        let errorMsg = "包含查询条件的情况下，查询时间段不允许超过15天。";
        if (queryLower.indexOf("like") > 0) {
          maxQueryRange = 3600 * 4 + 300;
          errorMsg = "查询条件中包含 like 的情况下，查询时间段不允许超过4小时。";
        }
        if(params.et - params.st > maxQueryRange) {
          message.error(errorMsg, 5);
          return;
        }
      }
    }
    return request<API.Res<LogsResponse>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tableId}/logs`,
      {
        cancelToken,
        method: "GET",
        params,
        skipErrorHandler: true,
      }
    );
  },

  // Get a list of log stores
  async getTableList(did: number) {
    return request<API.Res<TablesResponse[]>>(
      process.env.PUBLIC_PATH + `api/v1/databases/${did}/tables`,
      {
        method: "GET",
      }
    );
  },

  // Get local database and table
  async getLocalDatabasesAndTables(iid: number) {
    return request<API.Res<LocalTables[]>>(
      process.env.PUBLIC_PATH + `api/v1/instances/${iid}/databases-exist`,
      {
        method: "GET",
      }
    );
  },

  // Get local table columns
  async getTableColumns(iid: number, params: TableColumnsRequest) {
    return request<API.Res<TableColumnsResponse>>(
      process.env.PUBLIC_PATH + `api/v1/instances/${iid}/columns-self-built`,
      {
        method: "GET",
        params,
      }
    );
  },

  // Create a log library
  // async createdTable(did: number, data: CreatedLogLibraryRequest) {
  //   return request<API.Res<string>>(
  //     process.env.PUBLIC_PATH + `api/v1/databases/${did}/tables`,
  //     {
  //       method: "POST",
  //       data,
  //     }
  //   );
  // },

  // Create a log library V2
  async createdTableEachRow(data: CreatedLogLibraryRequestEachRow) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v2/storage`,
      {
        method: "POST",
        data,
      }
    );
  },

  // Create a log library V3
  async createdTableAsString(data: CreatedLogLibraryRequestAsString) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v3/storage`,
      {
        method: "POST",
        data,
      }
    );
  },

  // Create a log library Template
  async createdTableTemplate(template: string, data: CreatedTableTemplateType) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v2/storage/${template}`,
      {
        method: "POST",
        data,
      }
    );
  },

  async getMappingJson(data: { data: string }) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v2/storage/mapping-json`,
      {
        method: "POST",
        data,
      }
    );
  },

  async createdLocalTable(iid: number, data: CreateLocalTableRequest) {
    return request(
      process.env.PUBLIC_PATH + `api/v1/instances/${iid}/tables-exist`,
      {
        method: "POST",
        data,
      }
    );
  },

  async createdLocalTableBatch(
    iid: number,
    data: CreateLocalTableRequestBatch
  ) {
    return request(
      process.env.PUBLIC_PATH + `api/v1/instances/${iid}/tables-exist-batch`,
      {
        method: "POST",
        data,
      }
    );
  },

  // Deleting a Log Library
  async deletedTable(id: number) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${id}`,
      {
        method: "DELETE",
      }
    );
  },

  // Get log library details
  async getTableInfo(id: number) {
    return request<API.Res<TableInfoResponse>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${id}`,
      {
        method: "GET",
      }
    );
  },

  // Get log library details
  async updateTableInfo(id: number, data: UpdateTableInfoType) {
    return request<any>(process.env.PUBLIC_PATH + `api/v2/storage/${id}`, {
      method: "PATCH",
      data,
    });
  },

  // Obtain the table id from the third-party channel
  async getTableId(params: GetTableIdRequest) {
    return request<API.Res<number>>(
      process.env.PUBLIC_PATH + `api/v1/table/id`,
      {
        method: "GET",
        params,
      }
    );
  },

  // Get a list of databases
  async getDatabaseList(iid?: number) {
    return request<API.Res<DatabaseResponse[]>>(
      process.env.PUBLIC_PATH + `api/v1/instances/${iid || 0}/databases`,
      {
        method: "GET",
      }
    );
  },

  // Get index details
  async getIndexDetail(tid: number, id: number, params: IndexDetailRequest) {
    return request<API.Res<IndexDetail[]>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tid}/indexes/${id}`,
      {
        method: "GET",
        params,
      }
    );
  },

  // Add or modify index
  async setIndexes(tid: number, data: IndexRequest) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tid}/indexes`,
      {
        method: "PATCH",
        data,
      }
    );
  },

  // Get Index Edit List
  async getIndexes(tid: number) {
    return request<API.Res<IndexInfoType[]>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tid}/indexes`,
      {
        method: "GET",
      }
    );
  },

  // Obtain log configuration rules
  async getViews(tid: number) {
    return request<API.Res<ViewResponse[]>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tid}/views`,
      {
        method: "GET",
      }
    );
  },
  // Create a log configuration rule
  async createdView(tid: number, data: CreatedViewRequest) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v1/tables/${tid}/views`,
      {
        method: "POST",
        data,
      }
    );
  },

  // Update log configuration rules
  async updatedView(id: number, data: CreatedViewRequest) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v1/views/${id}`,
      {
        method: "PATCH",
        data,
      }
    );
  },

  async deletedView(id: number) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v1/views/${id}`,
      {
        method: "DELETE",
      }
    );
  },

  // Obtain rule details
  async getViewInfo(id: number) {
    return request<API.Res<ViewInfoResponse>>(
      process.env.PUBLIC_PATH + `api/v1/views/${id}`,
      {
        method: "GET",
      }
    );
  },

  // Obtain statistical tables
  async getStatisticalTable(iid: number, params: { query: string }) {
    return request<API.Res<StatisticalTableResponse>>(
      process.env.PUBLIC_PATH + `api/v1/instances/${iid}/complete`,
      {
        method: "POST",
        params,
      }
    );
  },

  // Hide Fields
  async getHideFields(tid: number) {
    return request<API.Res<any[]>>(
      process.env.PUBLIC_PATH + `api/v1/hidden/${tid}`,
      {
        method: "GET",
      }
    );
  },
  async updateHideFields(tid: number, data: { fields: string[] }) {
    return request<API.Res<string>>(
      process.env.PUBLIC_PATH + `api/v1/hidden/${tid}`,
      {
        method: "POST",
        data,
      }
    );
  },

  // Storage analysis field list
  async getAnalysisField(tid: number) {
    return request<API.Res<{ baseFields: any[]; logFields: any[] }>>(
      process.env.PUBLIC_PATH + `api/v2/storage/${tid}/analysis-fields`
    );
  },

  // 创建短链接
  async getShorturls(data: { originUrl: string }) {
    return request(process.env.PUBLIC_PATH + `api/v2/base/shorturls`, {
      method: "POST",
      data,
    });
  },

  // 链接链路日志库
  async updateLinkLinkLogLibrary(data: {
    storageId: number;
    traceTableId: number;
  }) {
    return request(
      process.env.PUBLIC_PATH + `api/v2/storage/${data.storageId}/trace`,
      {
        method: "PATCH",
        data: {
          traceTableId: data.traceTableId,
        },
      }
    );
  },

  // 链路日志库列表
  async getLinkLogLibraryList() {
    return request(process.env.PUBLIC_PATH + `api/v2/storage/traces`, {
      method: "GET",
    });
  },

  // 链路日志库依赖图
  async getLinkLogLibraryDependency(
    storageId: number,
    params?: { endTime?: number; startTime?: number }
  ) {
    return request(
      process.env.PUBLIC_PATH + `api/v2/storage/${storageId}/trace-graph`,
      {
        method: "GET",
        params,
      }
    );
  },

  // 获取日志filter列表
  async getLogFilterList(params: {
    collectType: CollectType;
    tableId?: number;
  }) {
    return request<API.Res<LogFilterType[]>>(
      process.env.PUBLIC_PATH + `api/v2/storage/collects`,
      {
        method: "GET",
        params,
      }
    );
  },

  // 创建日志filter
  async createLogFilter(data: {
    alias?: string;
    collectType: CollectType;
    statement: string;
    tableId?: number;
    column?: string; // 分析字段名称
  }) {
    return request(process.env.PUBLIC_PATH + `api/v2/storage/collects`, {
      method: "POST",
      data,
    });
  },

  // 删除日志filter
  async deleteLogFilter(collectId: number) {
    return request(
      process.env.PUBLIC_PATH + `api/v2/storage/collects/${collectId}`,
      {
        method: "DELETE",
      }
    );
  },

  // 编辑日志filter
  async editLogFilter(
    collectId: number,
    data: {
      alias?: string;
      collectType: CollectType;
      statement: string;
      tableId?: number;
      column?: string; // 分析字段名称
    }
  ) {
    return request(
      process.env.PUBLIC_PATH + `api/v2/storage/collects/${collectId}`,
      {
        method: "PATCH",
        data,
      }
    );
  },

  // 获取当前表的字段
  async getColumns(storageId: number) {
    return request(
      process.env.PUBLIC_PATH + `api/v2/storage/${storageId}/columns`,
      {
        method: "GET",
      }
    );
  },
};
