# ArangoDB Operator Metrics

## List

|                                                      Name                                                       |     Namespace     |    Group     | Type  | Description                                        |
|:---------------------------------------------------------------------------------------------------------------:|:-----------------:|:------------:|:-----:|:---------------------------------------------------|
|                     [arangodb_operator_agency_errors](./arangodb_operator_agency_errors.md)                     | arangodb_operator |    agency    | Count | Current count of agency cache fetch errors         |
|                    [arangodb_operator_agency_fetches](./arangodb_operator_agency_fetches.md)                    | arangodb_operator |    agency    | Count | Current count of agency cache fetches              |
|                      [arangodb_operator_agency_index](./arangodb_operator_agency_index.md)                      | arangodb_operator |    agency    | Gauge | Current index of the agency cache                  |
|       [arangodb_operator_agency_cache_health_present](./arangodb_operator_agency_cache_health_present.md)       | arangodb_operator | agency_cache | Gauge | Determines if local agency cache health is present |
|              [arangodb_operator_agency_cache_healthy](./arangodb_operator_agency_cache_healthy.md)              | arangodb_operator | agency_cache | Gauge | Determines if agency is healthy                    |
|              [arangodb_operator_agency_cache_leaders](./arangodb_operator_agency_cache_leaders.md)              | arangodb_operator | agency_cache | Gauge | Determines agency leader vote count                |
| [arangodb_operator_agency_cache_member_commit_offset](./arangodb_operator_agency_cache_member_commit_offset.md) | arangodb_operator | agency_cache | Gauge | Determines agency member commit offset             |
|       [arangodb_operator_agency_cache_member_serving](./arangodb_operator_agency_cache_member_serving.md)       | arangodb_operator | agency_cache | Gauge | Determines if agency member is reachable           |
|              [arangodb_operator_agency_cache_present](./arangodb_operator_agency_cache_present.md)              | arangodb_operator | agency_cache | Gauge | Determines if local agency cache is present        |
|              [arangodb_operator_agency_cache_serving](./arangodb_operator_agency_cache_serving.md)              | arangodb_operator | agency_cache | Gauge | Determines if agency is serving                    |
