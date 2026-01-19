## @yourproject/api-client@1.0

This generator creates TypeScript/JavaScript client that utilizes [axios](https://github.com/axios/axios). The generated Node module can be used in the following environments:

Environment
* Node.js
* Webpack
* Browserify

Language level
* ES5 - you must have a Promises/A+ library installed
* ES6

Module system
* CommonJS
* ES6 module system

It can be used in both TypeScript and JavaScript. In TypeScript, the definition will be automatically resolved via `package.json`. ([Reference](https://www.typescriptlang.org/docs/handbook/declaration-files/consumption.html))

### Building

To build and compile the typescript sources to javascript use:
```
npm install
npm run build
```

### Publishing

First build the package then run `npm publish`

### Consuming

navigate to the folder of your consuming project and run one of the following commands.

_published:_

```
npm install @yourproject/api-client@1.0 --save
```

_unPublished (not recommended):_

```
npm install PATH_TO_GENERATED_PACKAGE --save
```

### Documentation for API Endpoints

All URIs are relative to *http://localhost:23347*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*DefaultApi* | [**apiBackupCdnPost**](docs/DefaultApi.md#apibackupcdnpost) | **POST** /api/backup/cdn | 手动触发CDN增量备份
*DefaultApi* | [**apiBackupDatabasePost**](docs/DefaultApi.md#apibackupdatabasepost) | **POST** /api/backup/database | 手动触发数据库备份
*DefaultApi* | [**apiBackupHistoryGet**](docs/DefaultApi.md#apibackuphistoryget) | **GET** /api/backup/history | 获取备份历史记录
*DefaultApi* | [**apiBackupRestoreCdnBackupIdPost**](docs/DefaultApi.md#apibackuprestorecdnbackupidpost) | **POST** /api/backup/restore/cdn/{backup_id} | 从备份恢复CDN文件
*DefaultApi* | [**apiBackupStatusGet**](docs/DefaultApi.md#apibackupstatusget) | **GET** /api/backup/status | 获取备份系统状态
*DefaultApi* | [**apiBackupTriggerPost**](docs/DefaultApi.md#apibackuptriggerpost) | **POST** /api/backup/trigger | 手动触发全量备份
*DefaultApi* | [**apiPingGet**](docs/DefaultApi.md#apipingget) | **GET** /api/ping | 健康检查
*DefaultApi* | [**apiResourcesGet**](docs/DefaultApi.md#apiresourcesget) | **GET** /api/resources | 获取资源文件列表
*DefaultApi* | [**apiResourcesIdDelete**](docs/DefaultApi.md#apiresourcesiddelete) | **DELETE** /api/resources/{id} | 删除资源文件
*DefaultApi* | [**apiResourcesIdGet**](docs/DefaultApi.md#apiresourcesidget) | **GET** /api/resources/{id} | 获取资源文件
*DefaultApi* | [**apiResourcesUploadPost**](docs/DefaultApi.md#apiresourcesuploadpost) | **POST** /api/resources/upload | 上传资源文件
*DefaultApi* | [**apiSecurityBlockIpPost**](docs/DefaultApi.md#apisecurityblockippost) | **POST** /api/security/block/{ip} | 手动封禁IP地址
*DefaultApi* | [**apiSecurityBlockedIpsGet**](docs/DefaultApi.md#apisecurityblockedipsget) | **GET** /api/security/blocked-ips | 获取被封禁IP列表
*DefaultApi* | [**apiSecurityConnectionsGet**](docs/DefaultApi.md#apisecurityconnectionsget) | **GET** /api/security/connections | 获取实时连接统计
*DefaultApi* | [**apiSecurityIpStatsGet**](docs/DefaultApi.md#apisecurityipstatsget) | **GET** /api/security/ip-stats | 获取IP访问统计
*DefaultApi* | [**apiSecurityStatusGet**](docs/DefaultApi.md#apisecuritystatusget) | **GET** /api/security/status | 获取系统安全状态
*DefaultApi* | [**apiSecurityUnblockIpPost**](docs/DefaultApi.md#apisecurityunblockippost) | **POST** /api/security/unblock/{ip} | 手动解封IP地址
*DefaultApi* | [**apiSecurityWhitelistIpDelete**](docs/DefaultApi.md#apisecuritywhitelistipdelete) | **DELETE** /api/security/whitelist/{ip} | 从白名单移除IP
*DefaultApi* | [**apiSecurityWhitelistIpPost**](docs/DefaultApi.md#apisecuritywhitelistippost) | **POST** /api/security/whitelist/{ip} | 添加IP到白名单


### Documentation For Models

 - [ApiBackupHistoryGet200Response](docs/ApiBackupHistoryGet200Response.md)
 - [ApiBackupStatusGet200Response](docs/ApiBackupStatusGet200Response.md)
 - [ApiResourceFile](docs/ApiResourceFile.md)
 - [ApiResourceListResponse](docs/ApiResourceListResponse.md)
 - [ApiResourcesGet200Response](docs/ApiResourcesGet200Response.md)
 - [ApiResourcesUploadPost200Response](docs/ApiResourcesUploadPost200Response.md)
 - [ApiSecurityBlockIpPost200Response](docs/ApiSecurityBlockIpPost200Response.md)
 - [ApiSecurityBlockedIpsGet200Response](docs/ApiSecurityBlockedIpsGet200Response.md)
 - [ApiSecurityConnectionsGet200Response](docs/ApiSecurityConnectionsGet200Response.md)
 - [ApiSecurityIpStatsGet200Response](docs/ApiSecurityIpStatsGet200Response.md)
 - [ApiSecurityStatusGet200Response](docs/ApiSecurityStatusGet200Response.md)
 - [ApiUploadResourceResponse](docs/ApiUploadResourceResponse.md)
 - [ControllersBackupHistoryResponse](docs/ControllersBackupHistoryResponse.md)
 - [ControllersBackupRecord](docs/ControllersBackupRecord.md)
 - [ControllersBackupStatus](docs/ControllersBackupStatus.md)
 - [ControllersBlockedIP](docs/ControllersBlockedIP.md)
 - [ControllersConnectionStats](docs/ControllersConnectionStats.md)
 - [ControllersIPStats](docs/ControllersIPStats.md)
 - [ControllersSecurityStatus](docs/ControllersSecurityStatus.md)
 - [ResponseResponse](docs/ResponseResponse.md)
 - [ResponseResponseCode](docs/ResponseResponseCode.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization

Endpoints do not require authorization.

