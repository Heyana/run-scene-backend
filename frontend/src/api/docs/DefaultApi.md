# DefaultApi

All URIs are relative to *http://localhost:23347*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**apiBackupCdnPost**](#apibackupcdnpost) | **POST** /api/backup/cdn | 手动触发CDN增量备份|
|[**apiBackupDatabasePost**](#apibackupdatabasepost) | **POST** /api/backup/database | 手动触发数据库备份|
|[**apiBackupHistoryGet**](#apibackuphistoryget) | **GET** /api/backup/history | 获取备份历史记录|
|[**apiBackupRestoreCdnBackupIdPost**](#apibackuprestorecdnbackupidpost) | **POST** /api/backup/restore/cdn/{backup_id} | 从备份恢复CDN文件|
|[**apiBackupStatusGet**](#apibackupstatusget) | **GET** /api/backup/status | 获取备份系统状态|
|[**apiBackupTriggerPost**](#apibackuptriggerpost) | **POST** /api/backup/trigger | 手动触发全量备份|
|[**apiPingGet**](#apipingget) | **GET** /api/ping | 健康检查|
|[**apiResourcesGet**](#apiresourcesget) | **GET** /api/resources | 获取资源文件列表|
|[**apiResourcesIdDelete**](#apiresourcesiddelete) | **DELETE** /api/resources/{id} | 删除资源文件|
|[**apiResourcesIdGet**](#apiresourcesidget) | **GET** /api/resources/{id} | 获取资源文件|
|[**apiResourcesUploadPost**](#apiresourcesuploadpost) | **POST** /api/resources/upload | 上传资源文件|
|[**apiSecurityBlockIpPost**](#apisecurityblockippost) | **POST** /api/security/block/{ip} | 手动封禁IP地址|
|[**apiSecurityBlockedIpsGet**](#apisecurityblockedipsget) | **GET** /api/security/blocked-ips | 获取被封禁IP列表|
|[**apiSecurityConnectionsGet**](#apisecurityconnectionsget) | **GET** /api/security/connections | 获取实时连接统计|
|[**apiSecurityIpStatsGet**](#apisecurityipstatsget) | **GET** /api/security/ip-stats | 获取IP访问统计|
|[**apiSecurityStatusGet**](#apisecuritystatusget) | **GET** /api/security/status | 获取系统安全状态|
|[**apiSecurityUnblockIpPost**](#apisecurityunblockippost) | **POST** /api/security/unblock/{ip} | 手动解封IP地址|
|[**apiSecurityWhitelistIpDelete**](#apisecuritywhitelistipdelete) | **DELETE** /api/security/whitelist/{ip} | 从白名单移除IP|
|[**apiSecurityWhitelistIpPost**](#apisecuritywhitelistippost) | **POST** /api/security/whitelist/{ip} | 添加IP到白名单|

# **apiBackupCdnPost**
> ResponseResponse apiBackupCdnPost()

单独执行CDN文件的增量备份操作

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiBackupCdnPost();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | CDN备份任务已触发 |  -  |
|**500** | 触发失败或调度器未初始化 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiBackupDatabasePost**
> ResponseResponse apiBackupDatabasePost()

单独执行数据库备份操作，不包括CDN文件

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiBackupDatabasePost();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 数据库备份任务已触发 |  -  |
|**500** | 触发失败或调度器未初始化 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiBackupHistoryGet**
> ApiBackupHistoryGet200Response apiBackupHistoryGet()

获取系统所有备份操作的历史记录，包括成功和失败的备份

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let page: number; //页码，默认1 (optional) (default to 1)
let limit: number; //每页数量，默认20，最大100 (optional) (default to 20)
let type: string; //备份类型筛选：database | cdn | full (optional) (default to undefined)

const { status, data } = await apiInstance.apiBackupHistoryGet(
    page,
    limit,
    type
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | 页码，默认1 | (optional) defaults to 1|
| **limit** | [**number**] | 每页数量，默认20，最大100 | (optional) defaults to 20|
| **type** | [**string**] | 备份类型筛选：database | cdn | full | (optional) defaults to undefined|


### Return type

**ApiBackupHistoryGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 获取失败或调度器未初始化 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiBackupRestoreCdnBackupIdPost**
> ResponseResponse apiBackupRestoreCdnBackupIdPost()

使用指定的备份记录恢复CDN文件到指定状态（谨慎操作）

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let backupId: number; //备份记录ID（示例: 1） (default to undefined)

const { status, data } = await apiInstance.apiBackupRestoreCdnBackupIdPost(
    backupId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **backupId** | [**number**] | 备份记录ID（示例: 1） | defaults to undefined|


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | CDN文件恢复成功 |  -  |
|**400** | 无效的备份 ID |  -  |
|**404** | 备份记录不存在 |  -  |
|**500** | 恢复失败或调度器未初始化 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiBackupStatusGet**
> ApiBackupStatusGet200Response apiBackupStatusGet()

获取当前备份调度器和备份服务的运行状态信息

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiBackupStatusGet();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ApiBackupStatusGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 备份调度器未初始化 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiBackupTriggerPost**
> ResponseResponse apiBackupTriggerPost()

立即执行一次完整的系统备份，包括数据库和CDN文件

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiBackupTriggerPost();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 备份任务已触发 |  -  |
|**500** | 触发失败或调度器未初始化 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiPingGet**
> ResponseResponse apiPingGet()

检查API服务是否正常运行

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiPingGet();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 服务正常 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiResourcesGet**
> ApiResourcesGet200Response apiResourcesGet()

获取系统中所有资源文件的列表，支持分页和类型筛选

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let page: number; //页码，默认1 (optional) (default to 1)
let pageSize: number; //每页数量，默认10，最大100 (optional) (default to 10)
let type: string; //资源类型筛选：image | video | audio | document (optional) (default to undefined)
let libraryId: number; //媒体库ID筛选 (optional) (default to undefined)

const { status, data } = await apiInstance.apiResourcesGet(
    page,
    pageSize,
    type,
    libraryId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | 页码，默认1 | (optional) defaults to 1|
| **pageSize** | [**number**] | 每页数量，默认10，最大100 | (optional) defaults to 10|
| **type** | [**string**] | 资源类型筛选：image | video | audio | document | (optional) defaults to undefined|
| **libraryId** | [**number**] | 媒体库ID筛选 | (optional) defaults to undefined|


### Return type

**ApiResourcesGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 查询失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiResourcesIdDelete**
> ResponseResponse apiResourcesIdDelete()

永久删除指定的资源文件，包括本地和CDN存储

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; //资源ID（示例: 1） (default to undefined)

const { status, data } = await apiInstance.apiResourcesIdDelete(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | 资源ID（示例: 1） | defaults to undefined|


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 删除成功 |  -  |
|**400** | 无效的资源ID |  -  |
|**404** | 资源不存在 |  -  |
|**500** | 删除失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiResourcesIdGet**
> File apiResourcesIdGet()

根据ID获取资源文件，返回文件内容或重定向到CDN地址

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let id: string; //资源ID（示例: 1） (default to undefined)

const { status, data } = await apiInstance.apiResourcesIdGet(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | 资源ID（示例: 1） | defaults to undefined|


### Return type

**File**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, image/*, text/*


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 资源文件内容 |  -  |
|**302** | CDN重定向 |  -  |
|**400** | 无效的资源ID |  -  |
|**404** | 资源不存在或内容不可用 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiResourcesUploadPost**
> ApiResourcesUploadPost200Response apiResourcesUploadPost()

上传各类资源文件（图片、文档、音频、视频等）到系统

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let file: File; //资源文件（最大10MB） (default to undefined)
let libraryId: number; //媒体库ID，可选 (optional) (default to undefined)
let resourceType: string; //资源类型（系统自动识别） (optional) (default to undefined)

const { status, data } = await apiInstance.apiResourcesUploadPost(
    file,
    libraryId,
    resourceType
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **file** | [**File**] | 资源文件（最大10MB） | defaults to undefined|
| **libraryId** | [**number**] | 媒体库ID，可选 | (optional) defaults to undefined|
| **resourceType** | [**string**] | 资源类型（系统自动识别） | (optional) defaults to undefined|


### Return type

**ApiResourcesUploadPost200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 上传成功 |  -  |
|**400** | 文件上传失败或文件过大 |  -  |
|**500** | 保存失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityBlockIpPost**
> ApiSecurityBlockIpPost200Response apiSecurityBlockIpPost()

将指定IP地址添加到黑名单，阻止其访问系统

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let ip: string; //IP地址（示例: 192.168.1.100） (default to undefined)
let duration: number; //封禁时长(秒)，默认3600 (optional) (default to 3600)
let reason: string; //封禁原因 (optional) (default to '\"手动封禁\"')

const { status, data } = await apiInstance.apiSecurityBlockIpPost(
    ip,
    duration,
    reason
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **ip** | [**string**] | IP地址（示例: 192.168.1.100） | defaults to undefined|
| **duration** | [**number**] | 封禁时长(秒)，默认3600 | (optional) defaults to 3600|
| **reason** | [**string**] | 封禁原因 | (optional) defaults to '\"手动封禁\"'|


### Return type

**ApiSecurityBlockIpPost200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 封禁成功 |  -  |
|**400** | IP地址格式错误或已被封禁 |  -  |
|**500** | 封禁失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityBlockedIpsGet**
> ApiSecurityBlockedIpsGet200Response apiSecurityBlockedIpsGet()

获取当前被封禁的IP地址列表，包括封禁原因和到期时间

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let page: number; //页码，默认1 (optional) (default to 1)
let limit: number; //每页数量，默认20 (optional) (default to 20)

const { status, data } = await apiInstance.apiSecurityBlockedIpsGet(
    page,
    limit
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | 页码，默认1 | (optional) defaults to 1|
| **limit** | [**number**] | 每页数量，默认20 | (optional) defaults to 20|


### Return type

**ApiSecurityBlockedIpsGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 获取失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityConnectionsGet**
> ApiSecurityConnectionsGet200Response apiSecurityConnectionsGet()

获取当前活跃连接统计信息，用于DDoS监控和流量分析

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiSecurityConnectionsGet();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ApiSecurityConnectionsGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 获取失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityIpStatsGet**
> ApiSecurityIpStatsGet200Response apiSecurityIpStatsGet()

获取各IP地址的访问次数、最后访问时间等统计信息

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let page: number; //页码，默认1 (optional) (default to 1)
let limit: number; //每页数量，默认50 (optional) (default to 50)
let sort: string; //排序方式：request_count(请求数) | last_access(最后访问) (optional) (default to '\"request_count\"')

const { status, data } = await apiInstance.apiSecurityIpStatsGet(
    page,
    limit,
    sort
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | 页码，默认1 | (optional) defaults to 1|
| **limit** | [**number**] | 每页数量，默认50 | (optional) defaults to 50|
| **sort** | [**string**] | 排序方式：request_count(请求数) | last_access(最后访问) | (optional) defaults to '\"request_count\"'|


### Return type

**ApiSecurityIpStatsGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 获取失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityStatusGet**
> ApiSecurityStatusGet200Response apiSecurityStatusGet()

获取当前安全中间件运行状态、配置信息和威胁等级

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.apiSecurityStatusGet();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ApiSecurityStatusGet200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 获取成功 |  -  |
|**500** | 获取失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityUnblockIpPost**
> ResponseResponse apiSecurityUnblockIpPost()

将指定IP地址从黑名单中移除，恢复其正常访问权限

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let ip: string; //IP地址（示例: 192.168.1.100） (default to undefined)

const { status, data } = await apiInstance.apiSecurityUnblockIpPost(
    ip
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **ip** | [**string**] | IP地址（示例: 192.168.1.100） | defaults to undefined|


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 解封成功 |  -  |
|**400** | IP地址格式错误 |  -  |
|**404** | IP未被封禁 |  -  |
|**500** | 解封失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityWhitelistIpDelete**
> ResponseResponse apiSecurityWhitelistIpDelete()

将IP地址从白名单中移除，该IP将重新受到安全检查

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let ip: string; //IP地址（示例: 192.168.1.1） (default to undefined)

const { status, data } = await apiInstance.apiSecurityWhitelistIpDelete(
    ip
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **ip** | [**string**] | IP地址（示例: 192.168.1.1） | defaults to undefined|


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 移除成功 |  -  |
|**400** | IP地址格式错误 |  -  |
|**404** | IP不在白名单中 |  -  |
|**500** | 移除失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiSecurityWhitelistIpPost**
> ResponseResponse apiSecurityWhitelistIpPost()

将IP地址添加到白名单，该IP将绕过所有安全检查

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@yourproject/api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let ip: string; //IP地址（示例: 192.168.1.1） (default to undefined)

const { status, data } = await apiInstance.apiSecurityWhitelistIpPost(
    ip
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **ip** | [**string**] | IP地址（示例: 192.168.1.1） | defaults to undefined|


### Return type

**ResponseResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | 添加成功 |  -  |
|**400** | IP地址格式错误或已在白名单 |  -  |
|**500** | 添加失败 |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

