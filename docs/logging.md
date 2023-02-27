## logging SCOPE

the form to show the code you need check logs up

### english

|  key word  | meaning  | 类型|
|  ----     | ----  | ---- |
| BUCKET_NUMBER  | bucket number  | int|
| BUCKET_SIZE | bucket size | int|
| BUCKET_PARALLEL | bucket goroutine pool| int|
| BUCKET_ONLINE    | bucket online  | int|
| BUCKET_ID    | bucket id |int|
| MONITOR_ONLINE_INTERVAL    | print online interval  |int|
| ERROR     | error  | string |
| ONLINE    | print online users number  |int|
| PPROF    | pprof url |string|
| ID    | user identification |string|
| FINISH_TASK    | finish task | string|
| PANIC    | panic | string|
| SPEND_TIME    | the spend time of  write message to connection |string|
| WEAK_NET    | users' connection is in weak status |string|
| OFFLINE_CAUSE    | connection closed by CAUSE | string|
|COUNT_LOSE_CONTENT|The data packets lost during the period between the last printing and the current printing |int64|
|COUNT_CONTENT|The total data packets during the period between the last printing and the current printing|int64|
|COUNT_CONTENT_LEN|The total packet length during the period between the last printing and the current printing|int64|

### chinese

|  key word  | meaning  | 类型|
|  ----     | ----  | ---- |
| BUCKET_NUMBER  | 配置的Bucket数量 | int|
| BUCKET_SIZE | 配置的Bucket的尺寸| int|
| BUCKET_PARALLEL | bucket并发线程数量| int|
| BUCKET_ONLINE    | 当前bucket在线人数 | int|
| BUCKET_ID    | 桶ID |int|
| MONITOR_ONLINE_INTERVAL    | 监控打印的时间间隔 |int|
| ERROR     | 错误 | string |
| ONLINE    | 当前所有用户在线数量 |int|
| PPROF    | pprof访问地址 |string|
| ID    | 用户ID |string|
| FINISH_TASK    | 结束的异步任务 | string|
| PANIC    | 程序出现panic被捕获 | string|
| SPEND_TIME    | 用户下推失败，出现丢包 |string|
| WEAK_NET    | 用户链接写入时间过长：最终还是写入成功 |string|
| OFFLINE_CAUSE    | 用户连接断线的原因 | string|
|COUNT_LOSE_CONTENT|上次打印丢包到这次打印丢包期间丢包总数量|int64|
|COUNT_CONTENT|从上次打印下推消息到这次下推消息总数量（包数量）|int64|
|COUNT_CONTENT_LEN|从上一次下推到现在总的下推消息的总长度|int64|

### monitor

It is recommended to aggregate and monitor the following 
fields, usually there are index requirements for the following fields

- Panic
- Error
- Online
- Bucket
- CountLoseContent
- CountContent
- CountContentLen






