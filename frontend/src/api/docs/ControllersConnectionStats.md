# ControllersConnectionStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**active_connections** | **number** |  | [optional] [default to undefined]
**ip_connections** | **{ [key: string]: number; }** |  | [optional] [default to undefined]
**top_ips** | [**Array&lt;ControllersIPStats&gt;**](ControllersIPStats.md) |  | [optional] [default to undefined]
**total_connections** | **number** |  | [optional] [default to undefined]

## Example

```typescript
import { ControllersConnectionStats } from '@yourproject/api-client';

const instance: ControllersConnectionStats = {
    active_connections,
    ip_connections,
    top_ips,
    total_connections,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
