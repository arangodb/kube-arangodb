# Change Log

## [master](https://github.com/arangodb/kube-arangodb/tree/master) (N/A)
- Add Encryption Key rotation feature for ArangoDB EE 3.7+

## [1.0.3](https://github.com/arangodb/kube-arangodb/tree/1.0.3) (2020-05-25)
- Prevent deletion of not known PVC's
- Move Restore as Plan

## [1.0.2](https://github.com/arangodb/kube-arangodb/tree/1.0.2) (2020-04-16)
- Added additional checks in UpToDate condition
- Added extended Rotation check for Cluster mode
- Removed old rotation logic (rotation of ArangoDeployment may be enforced after Operator upgrade)
- Added UpToDate condition in ArangoDeployment Status

## [1.0.1](https://github.com/arangodb/kube-arangodb/tree/1.0.1) (2020-03-25)
- Added Customizable Affinity settings for ArangoDB Member Pods
- Added possibility to override default images used by ArangoDeployment
- Added possibility to set probes on all groups
- Added Image Discovery type in ArangoDeployment spec
- Prevent Agency Members recreation
- Added Customizable Volumes and VolumeMounts for ArangoDB server container
- Added MemoryOverride flag for ArangoDB >= 3.6.3
- Improved Rotation discovery process
- Added annotation to rotate ArangoDeployment in secure way

## [1.0.0](https://github.com/arangodb/kube-arangodb/tree/1.0.0) (2020-03-03)
- Removal of v1alpha support for ArangoDeployment, ArangoDeploymentReplication, ArangoBackup
- Added new command to operator - version

## [0.4.5](https://github.com/arangodb/kube-arangodb/tree/0.4.5) (2020-03-02)
- Add Customizable SecurityContext for ArangoDeployment pods

## [0.4.4](https://github.com/arangodb/kube-arangodb/tree/0.4.4) (2020-02-27)
- Add new VolumeResize mode to be compatible with Azure flow
- Allow to customize probe configuration options
- Add new upgrade flag for ArangoDB 3.6.0<=

## [0.4.3](https://github.com/arangodb/kube-arangodb/tree/0.4.3) (2020-01-31)
- Prevent DBServer deletion if there are any shards active on it
- Add Maintenance mode annotation for ArangoDeployment

## [0.4.2](https://github.com/arangodb/kube-arangodb/tree/0.4.2) (2019-11-12)
- AntiAffinity for operator pods.
- Add CRD API v1 with support for v1alpha.
- Allow to set annotations in ArangoDeployment resources.
- Add UBI based image.

## [0.4.0](https://github.com/arangodb/kube-arangodb/tree/0.4.0) (2019-10-09)
- Further helm chart fixes for linter.
- Support hot backup.
- Disable scaling buttons if scaling is not possible.

## [0.3.16](https://github.com/arangodb/kube-arangodb/tree/0.3.16) (2019-09-25)
- Revised helm charts.
- Use separate service account for operator.
- Support for ResignLeadership job.
- Allow to set ImagePullSecrets in pods.
- Bug fixes.

## [0.3.15]() (never released, only previews existed)

## [0.3.14](https://github.com/arangodb/kube-arangodb/tree/0.3.14) (2019-08-07)
- Bug fixes for custom sidecars.
- More tests

## [0.3.13](https://github.com/arangodb/kube-arangodb/tree/0.3.13) (2019-08-02)
- Added side car changed to pod rotation criterium
- Added ArangoDB version and image id to member status
- Fix bug with MemberOfCluster condition
- Added test for resource change

## [0.3.12](https://github.com/arangodb/kube-arangodb/tree/0.3.12) (2019-07-04)
- Limit source IP ranges for external services

## [0.3.11](https://github.com/arangodb/kube-arangodb/tree/0.3.11) (2019-06-07)
- Introduced volume claim templates for all server groups that require volume.
- Added arangodb-exporter support as sidecar to all arangodb pods.
- Fixed a bug in the case that all coordinators failed.
- Increase some timeouts in cluster observation.
- Ignore connection errors when removing servers.
- Switch to go 1.12 and modules.
- User sidecars.

## [0.3.10](https://github.com/arangodb/kube-arangodb/tree/0.3.10) (2019-04-04)
- Added Pod Disruption Budgets for all server groups in production mode.
- Added Priority Class Name to be specified per server group.
- Forward resource requirements to k8s.
- Automatic creation of randomized root password on demand.
- Volume resizing (only enlarge).
- Allow to disable liveness probes, increase timeouts in defaults.
- Handle case of all coordinators gone better.
- Added `MY_NODE_NAME` and `NODE_NAME` env vars for all pods.
- Internal communications with ArangoDB more secure through tokens which
  are limited to certain API paths.
- Rolling upgrade waits till all shards are in sync before proceeding to
  next dbserver, even if it takes longer than 15 min.
- Improve installation and upgrade instructions in README.

## [0.3.9](https://github.com/arangodb/kube-arangodb/tree/0.3.9) (2019-02-28)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.8...0.3.9)
- Fixed a serious bug in rolling upgrades which was introduced in 0.3.8.
- Document the drain procedure for k8s nodes.
- Wait for shards to be in sync before continuing upgrade process.
- Rotate members when patch-level upgrade.
- Don't trigger cleanout server during upgrade.
- More robust remove-server actions.

## [0.3.8](https://github.com/arangodb/kube-arangodb/tree/0.3.8) (2019-02-19)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.7...0.3.8)

- Added scaling limits to spec and enforce in operator.
- npm update for dashboard to alleviate security problems.
- Added bare metal walk through to documentation.
- Wait for coordinator to be ready in kubernetes.
- Schedule only one CleanOutServer job in drain scenario, introduce
  Drain phase.
- Take care of case that server is terminated by drain before cleanout
  has completed.
- Added undocumented force-status-reload status field.
- Take care of case that all coordinators have failed: delete all
  coordinator pods and create new ones.
- Updated lodash for dashboard.
- Try harder to remove server from cluster if it does not work right away.
- Update member status, if once decided to drain, continue draining.
  This takes care of more corner cases.

## [0.3.7](https://github.com/arangodb/kube-arangodb/tree/0.3.7) (2019-01-03)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.6...0.3.7)

**Merged pull requests:**

- Use jwt-keyfile option if available. [\#318](https://github.com/arangodb/kube-arangodb/pull/318)
- StorageOperator Volume Size Fix [\#316](https://github.com/arangodb/kube-arangodb/pull/316)

## [0.3.6](https://github.com/arangodb/kube-arangodb/tree/0.3.6) (2018-12-06)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.5...0.3.6)

**Closed issues:**

- Dashboards not aware of kube-proxy [\#278](https://github.com/arangodb/kube-arangodb/issues/278)

**Merged pull requests:**

- Link to k8s platform tutorials. [\#313](https://github.com/arangodb/kube-arangodb/pull/313)
- Updated Go-Driver to latest version. [\#312](https://github.com/arangodb/kube-arangodb/pull/312)
- NodeSelector [\#311](https://github.com/arangodb/kube-arangodb/pull/311)
- Docs: Formatting [\#310](https://github.com/arangodb/kube-arangodb/pull/310)
- Doc: remove duplicate chapter [\#309](https://github.com/arangodb/kube-arangodb/pull/309)
- Doc: remove blanks after tripple tics [\#308](https://github.com/arangodb/kube-arangodb/pull/308)
- License Key [\#307](https://github.com/arangodb/kube-arangodb/pull/307)
- Updated packages containing vulnerabilities [\#306](https://github.com/arangodb/kube-arangodb/pull/306)
- Advertised Endpoints [\#299](https://github.com/arangodb/kube-arangodb/pull/299)

## [0.3.5](https://github.com/arangodb/kube-arangodb/tree/0.3.5) (2018-11-20)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.4...0.3.5)

**Closed issues:**

- Istio compatibility issue [\#260](https://github.com/arangodb/kube-arangodb/issues/260)

**Merged pull requests:**

- Fixing imageID retrieval issue when sidecars are injected. [\#302](https://github.com/arangodb/kube-arangodb/pull/302)
- Bug fix/fix immutable reset [\#301](https://github.com/arangodb/kube-arangodb/pull/301)
- Fixing small type in readme [\#300](https://github.com/arangodb/kube-arangodb/pull/300)
- Make timeout configurable. [\#298](https://github.com/arangodb/kube-arangodb/pull/298)
- fixed getLoadBalancerIP to also handle hostnames [\#297](https://github.com/arangodb/kube-arangodb/pull/297)

## [0.3.4](https://github.com/arangodb/kube-arangodb/tree/0.3.4) (2018-11-06)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.3...0.3.4)

**Merged pull requests:**

- Try to repair changelog generator. [\#296](https://github.com/arangodb/kube-arangodb/pull/296)
- Fixing uninitialised `lastNumberOfServers`. [\#294](https://github.com/arangodb/kube-arangodb/pull/294)
- Fixes for semiautomation. [\#293](https://github.com/arangodb/kube-arangodb/pull/293)
- add ebs volumes to eks doc [\#295](https://github.com/arangodb/kube-arangodb/pull/295)

## [0.3.3](https://github.com/arangodb/kube-arangodb/tree/0.3.3) (2018-11-02)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.2...0.3.3)

**Closed issues:**

- `manifests/arango-crd.yaml` not in repository [\#292](https://github.com/arangodb/kube-arangodb/issues/292)

**Merged pull requests:**

- Make semiautomation files self-contained. [\#291](https://github.com/arangodb/kube-arangodb/pull/291)

## [0.3.2](https://github.com/arangodb/kube-arangodb/tree/0.3.2) (2018-11-02)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.1...0.3.2)

**Closed issues:**

- Operator redeployed not fully functional [\#273](https://github.com/arangodb/kube-arangodb/issues/273)
- Busy Update Loop on PKS [\#272](https://github.com/arangodb/kube-arangodb/issues/272)
- scaling down in production starts pending pods to terminate them immediately [\#267](https://github.com/arangodb/kube-arangodb/issues/267)
- crd inclusion in helm chart prevents subsequent deployments to alternate namespaces [\#261](https://github.com/arangodb/kube-arangodb/issues/261)
- Tutorials with real world examples [\#229](https://github.com/arangodb/kube-arangodb/issues/229)

**Merged pull requests:**

- UI Fix [\#290](https://github.com/arangodb/kube-arangodb/pull/290)
- Revisited scale up and scale down. [\#288](https://github.com/arangodb/kube-arangodb/pull/288)
- Bug fix/extra crd yaml [\#287](https://github.com/arangodb/kube-arangodb/pull/287)
- Documentation/add aks tutorial [\#286](https://github.com/arangodb/kube-arangodb/pull/286)
- IPv6 revisited [\#285](https://github.com/arangodb/kube-arangodb/pull/285)
- Bug fix/readiness upgrade fix [\#283](https://github.com/arangodb/kube-arangodb/pull/283)
- Revert "Skip LoadBalancer Test" [\#282](https://github.com/arangodb/kube-arangodb/pull/282)
- Updated node modules to fix vulnerabilities [\#281](https://github.com/arangodb/kube-arangodb/pull/281)
- First stab at semiautomation. [\#280](https://github.com/arangodb/kube-arangodb/pull/280)
- When doing tests, always pull the image. [\#279](https://github.com/arangodb/kube-arangodb/pull/279)
- Break PKS Loop [\#277](https://github.com/arangodb/kube-arangodb/pull/277)
- Fixed readiness route. [\#276](https://github.com/arangodb/kube-arangodb/pull/276)
- Bug fix/scale up error [\#275](https://github.com/arangodb/kube-arangodb/pull/275)
- minor fix in template generation [\#274](https://github.com/arangodb/kube-arangodb/pull/274)
- Added `disableIPV6` Spec entry. [\#271](https://github.com/arangodb/kube-arangodb/pull/271)
- Test Image Option [\#270](https://github.com/arangodb/kube-arangodb/pull/270)
- Skip LoadBalancer Test [\#269](https://github.com/arangodb/kube-arangodb/pull/269)
- Test/templates [\#266](https://github.com/arangodb/kube-arangodb/pull/266)
- Updated examples to use version 3.3.17. [\#265](https://github.com/arangodb/kube-arangodb/pull/265)
- Unified Readiness Test [\#264](https://github.com/arangodb/kube-arangodb/pull/264)
- Use correct templateoptions for helm charts [\#258](https://github.com/arangodb/kube-arangodb/pull/258)
- Add advanced dc2dc to acceptance test. [\#252](https://github.com/arangodb/kube-arangodb/pull/252)
- adding EKS tutorial [\#289](https://github.com/arangodb/kube-arangodb/pull/289)

## [0.3.1](https://github.com/arangodb/kube-arangodb/tree/0.3.1) (2018-09-25)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.3.0...0.3.1)

**Closed issues:**

- Helm chart not deploying custom resource definitions [\#254](https://github.com/arangodb/kube-arangodb/issues/254)
- `go get` failing due to nonexistent arangodb/arangosync repo [\#249](https://github.com/arangodb/kube-arangodb/issues/249)
- Helm chart download links broken \(404\) [\#248](https://github.com/arangodb/kube-arangodb/issues/248)
- Make it easy to deploy in another namespace [\#230](https://github.com/arangodb/kube-arangodb/issues/230)
- Deployment Failed to Start in different Namespace other than Default [\#223](https://github.com/arangodb/kube-arangodb/issues/223)

**Merged pull requests:**

- Bugfix/sed on linux [\#259](https://github.com/arangodb/kube-arangodb/pull/259)
- README updates, removing `kubectl apply -f crd.yaml` [\#256](https://github.com/arangodb/kube-arangodb/pull/256)
- Include CRD in helm chart [\#255](https://github.com/arangodb/kube-arangodb/pull/255)

## [0.3.0](https://github.com/arangodb/kube-arangodb/tree/0.3.0) (2018-09-07)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.2.2...0.3.0)

**Closed issues:**

- Provide an option to add SubjectAltName or option to disable SSL [\#239](https://github.com/arangodb/kube-arangodb/issues/239)
- Use go-upgrade-rules [\#234](https://github.com/arangodb/kube-arangodb/issues/234)
- Spot the difference [\#225](https://github.com/arangodb/kube-arangodb/issues/225)
- How to Delete ArangoDeployment [\#224](https://github.com/arangodb/kube-arangodb/issues/224)
- Unable to delete pods, stuck in terminating state [\#220](https://github.com/arangodb/kube-arangodb/issues/220)
- Do not allow "critical" cmdline arguments to be overwritten [\#207](https://github.com/arangodb/kube-arangodb/issues/207)

**Merged pull requests:**

- Avoid use of arangosync packages [\#250](https://github.com/arangodb/kube-arangodb/pull/250)
- Fixed PV creation on kubernetes 1.11 [\#247](https://github.com/arangodb/kube-arangodb/pull/247)
- Resilience improvements [\#246](https://github.com/arangodb/kube-arangodb/pull/246)
- Adding GKE tutorial [\#245](https://github.com/arangodb/kube-arangodb/pull/245)
- Reject critical options during validation fixes \#207 [\#243](https://github.com/arangodb/kube-arangodb/pull/243)
- Trying to stabalize resilience tests [\#242](https://github.com/arangodb/kube-arangodb/pull/242)
- Adding helm charts for deploying the operators [\#238](https://github.com/arangodb/kube-arangodb/pull/238)
- Include license in upgrade check [\#237](https://github.com/arangodb/kube-arangodb/pull/237)
- Use new CurrentImage field to prevent unintended upgrades. [\#236](https://github.com/arangodb/kube-arangodb/pull/236)
- Use go-upgrade-rules to make "is upgrade allowed" decision fixes \#234 [\#235](https://github.com/arangodb/kube-arangodb/pull/235)
- Updated versions to known "proper" versions [\#233](https://github.com/arangodb/kube-arangodb/pull/233)
- Applying defaults after immutable fields have been reset [\#232](https://github.com/arangodb/kube-arangodb/pull/232)
- Updated go-driver to latest version [\#231](https://github.com/arangodb/kube-arangodb/pull/231)
- EE note for Kubernetes DC2DC [\#222](https://github.com/arangodb/kube-arangodb/pull/222)
- Documented dashboard usage [\#219](https://github.com/arangodb/kube-arangodb/pull/219)
- Load balancing tests [\#218](https://github.com/arangodb/kube-arangodb/pull/218)
- Add links to other operators in dashboard menu [\#217](https://github.com/arangodb/kube-arangodb/pull/217)
- Grouping style elements in 1 place [\#216](https://github.com/arangodb/kube-arangodb/pull/216)
- Adding ArangoDeploymentReplication dashboard. [\#215](https://github.com/arangodb/kube-arangodb/pull/215)
- Do not build initcontainer for imageid pod [\#214](https://github.com/arangodb/kube-arangodb/pull/214)
- Dashboard for ArangoLocalStorage operator [\#213](https://github.com/arangodb/kube-arangodb/pull/213)
- Adjust documentation based on new load balancer support. [\#212](https://github.com/arangodb/kube-arangodb/pull/212)
- Feature/dashboard [\#211](https://github.com/arangodb/kube-arangodb/pull/211)
- Use gin as HTTP server framework [\#210](https://github.com/arangodb/kube-arangodb/pull/210)
- Dashboard design concept [\#209](https://github.com/arangodb/kube-arangodb/pull/209)

## [0.2.2](https://github.com/arangodb/kube-arangodb/tree/0.2.2) (2018-06-29)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.2.1...0.2.2)

**Closed issues:**

- Unable to unset standard storage class in GKE using kubectl [\#200](https://github.com/arangodb/kube-arangodb/issues/200)
- Fix operators Deployment spec wrt minimum availability [\#198](https://github.com/arangodb/kube-arangodb/issues/198)
- Rotate server when cmdline arguments change [\#189](https://github.com/arangodb/kube-arangodb/issues/189)

**Merged pull requests:**

- Set a `role=leader` label on the Pod who won the leader election [\#208](https://github.com/arangodb/kube-arangodb/pull/208)
- Rotate server on changed arguments [\#206](https://github.com/arangodb/kube-arangodb/pull/206)
- Documentation fixes [\#205](https://github.com/arangodb/kube-arangodb/pull/205)
- Fixed get/set Default flag for StorageClasses [\#204](https://github.com/arangodb/kube-arangodb/pull/204)
- Log improvements [\#203](https://github.com/arangodb/kube-arangodb/pull/203)
- All operator Pods will now reach the Ready state. [\#201](https://github.com/arangodb/kube-arangodb/pull/201)

## [0.2.1](https://github.com/arangodb/kube-arangodb/tree/0.2.1) (2018-06-19)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.2.0...0.2.1)

## [0.2.0](https://github.com/arangodb/kube-arangodb/tree/0.2.0) (2018-06-19)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.1.0...0.2.0)

**Closed issues:**

- Guard operations that yield downtime with an `downtimeAllowed` field [\#190](https://github.com/arangodb/kube-arangodb/issues/190)
- Require at least 2 dbservers for `Cluster` deployment [\#178](https://github.com/arangodb/kube-arangodb/issues/178)
- Resource re-deployments when changing specific specs [\#164](https://github.com/arangodb/kube-arangodb/issues/164)
- PVC's can get stuck in Terminating state [\#157](https://github.com/arangodb/kube-arangodb/issues/157)
- PVC [\#156](https://github.com/arangodb/kube-arangodb/issues/156)
- Add timeout for reconciliation plan\(items\) [\#154](https://github.com/arangodb/kube-arangodb/issues/154)
- Add setting to specify ServiceAccount for deployment [\#146](https://github.com/arangodb/kube-arangodb/issues/146)
- Finalizers TODO [\#138](https://github.com/arangodb/kube-arangodb/issues/138)
- Prevent deleting pods \(manually\) using finalizers [\#134](https://github.com/arangodb/kube-arangodb/issues/134)
- Set controller of pods to support `kubectl drain` [\#132](https://github.com/arangodb/kube-arangodb/issues/132)
- Add option to taint pods [\#131](https://github.com/arangodb/kube-arangodb/issues/131)
- OpenShift: No DB is getting deployed [\#128](https://github.com/arangodb/kube-arangodb/issues/128)
- ArangoDeploymentTasks [\#34](https://github.com/arangodb/kube-arangodb/issues/34)
- ArangoLocalStorage tasks [\#33](https://github.com/arangodb/kube-arangodb/issues/33)

**Merged pull requests:**

- Adding downtimeAllowed field [\#194](https://github.com/arangodb/kube-arangodb/pull/194)
- Added tutorial for configuring DC2DC of Kubernetes [\#187](https://github.com/arangodb/kube-arangodb/pull/187)
- Various TLS & Sync related fixes [\#186](https://github.com/arangodb/kube-arangodb/pull/186)
- Use standard EventRecord to use event compression [\#185](https://github.com/arangodb/kube-arangodb/pull/185)
- Fixed ID prefix for single servers [\#184](https://github.com/arangodb/kube-arangodb/pull/184)
- Allow changing server group storage class. [\#183](https://github.com/arangodb/kube-arangodb/pull/183)
- Added test timeouts to all stages [\#182](https://github.com/arangodb/kube-arangodb/pull/182)
- Added renewal of deployment TLS CA certificate [\#181](https://github.com/arangodb/kube-arangodb/pull/181)
- Min dbserver count is 2. Revert phase when cleanout has failed [\#180](https://github.com/arangodb/kube-arangodb/pull/180)
- Prefer distinct nodes, even when not required [\#179](https://github.com/arangodb/kube-arangodb/pull/179)
- Added duration test app [\#177](https://github.com/arangodb/kube-arangodb/pull/177)
- Improved readiness probe, database services only use ready pods [\#176](https://github.com/arangodb/kube-arangodb/pull/176)
- Documenting acceptance test [\#175](https://github.com/arangodb/kube-arangodb/pull/175)
- Avoid useless warnings in log [\#174](https://github.com/arangodb/kube-arangodb/pull/174)
- Hide "dangerous" functions of MemberStatusList [\#173](https://github.com/arangodb/kube-arangodb/pull/173)
- Avoid overwriting status changes [\#172](https://github.com/arangodb/kube-arangodb/pull/172)
- Abort reconcilientation plan on failed cleanout server [\#171](https://github.com/arangodb/kube-arangodb/pull/171)
- Improving documentation [\#170](https://github.com/arangodb/kube-arangodb/pull/170)
- Remove service stickyness [\#169](https://github.com/arangodb/kube-arangodb/pull/169)
- Prevent deleting the PV when the PVC has already been attached to it [\#168](https://github.com/arangodb/kube-arangodb/pull/168)
- Various test improvements [\#167](https://github.com/arangodb/kube-arangodb/pull/167)
- Added unit tests for pv\_creator.go [\#166](https://github.com/arangodb/kube-arangodb/pull/166)
- Added finalizer on deployment, used to remove child finalizers on delete [\#165](https://github.com/arangodb/kube-arangodb/pull/165)
- Fix endless rotation because of serviceAccount `default` [\#163](https://github.com/arangodb/kube-arangodb/pull/163)
- Force volumes to unique nodes for production environments [\#162](https://github.com/arangodb/kube-arangodb/pull/162)
- Improved Service documentation [\#161](https://github.com/arangodb/kube-arangodb/pull/161)
- Reconciliation plan-item timeout [\#160](https://github.com/arangodb/kube-arangodb/pull/160)
- Operator high-availability [\#155](https://github.com/arangodb/kube-arangodb/pull/155)
- Cleanup long terminating stateful pods [\#153](https://github.com/arangodb/kube-arangodb/pull/153)
- Allow customization of serviceAccountName for pods [\#152](https://github.com/arangodb/kube-arangodb/pull/152)
- Cleanup stateless pods that are in terminating state for a long time [\#151](https://github.com/arangodb/kube-arangodb/pull/151)
- Added no-execute tolerations on operators to failover quicker [\#150](https://github.com/arangodb/kube-arangodb/pull/150)
- Replication shard status in ArangoDeploymentReplication status [\#148](https://github.com/arangodb/kube-arangodb/pull/148)
- Sync access packages [\#147](https://github.com/arangodb/kube-arangodb/pull/147)
- Adding syncmaster&worker reconciliation support. [\#145](https://github.com/arangodb/kube-arangodb/pull/145)
- Fixes needed to run on latest openshift. [\#144](https://github.com/arangodb/kube-arangodb/pull/144)
- `ArangoDeploymentReplication` resource [\#143](https://github.com/arangodb/kube-arangodb/pull/143)
- Adding deployment replication spec [\#142](https://github.com/arangodb/kube-arangodb/pull/142)
- No stickyness for EA service of type LoadBalancer [\#141](https://github.com/arangodb/kube-arangodb/pull/141)
- Added `tolerations` field to configure tolerations of generated pods. [\#140](https://github.com/arangodb/kube-arangodb/pull/140)
- Inspect node schedulable state [\#139](https://github.com/arangodb/kube-arangodb/pull/139)
- Make use of GOCACHE as docker volume for improved build times [\#137](https://github.com/arangodb/kube-arangodb/pull/137)
- Feature: finalizers [\#136](https://github.com/arangodb/kube-arangodb/pull/136)
- Added a spec regarding the rules for eviction & replacement of pods [\#133](https://github.com/arangodb/kube-arangodb/pull/133)
- Added support for running arangosync master & worker servers. [\#130](https://github.com/arangodb/kube-arangodb/pull/130)
- Updated go-certificates & go-driver to latest versions [\#127](https://github.com/arangodb/kube-arangodb/pull/127)
- Added Database external access service feature [\#126](https://github.com/arangodb/kube-arangodb/pull/126)
- Updated to latest go-driver [\#125](https://github.com/arangodb/kube-arangodb/pull/125)
- BREAKING CHANGE: Deployment mode ResilientSingle renamed to ActiveFailover [\#124](https://github.com/arangodb/kube-arangodb/pull/124)
- add persistent-volume tests [\#97](https://github.com/arangodb/kube-arangodb/pull/97)

## [0.1.0](https://github.com/arangodb/kube-arangodb/tree/0.1.0) (2018-04-06)
[Full Changelog](https://github.com/arangodb/kube-arangodb/compare/0.0.1...0.1.0)

**Closed issues:**

- make sure scripts terminate to avoid hanging CI [\#63](https://github.com/arangodb/kube-arangodb/issues/63)
- prefix environment variables [\#62](https://github.com/arangodb/kube-arangodb/issues/62)
- warning when passing string literal "None" as spec.tls.caSecretName [\#60](https://github.com/arangodb/kube-arangodb/issues/60)

**Merged pull requests:**

- Fixed down/upgrading resilient single deployments. [\#123](https://github.com/arangodb/kube-arangodb/pull/123)
- Various docs improvements & fixes [\#122](https://github.com/arangodb/kube-arangodb/pull/122)
- Added tests for query cursors on various deployments. [\#121](https://github.com/arangodb/kube-arangodb/pull/121)
- Remove upgrade resilient single 3.2 -\> 3.3 test. [\#120](https://github.com/arangodb/kube-arangodb/pull/120)
- Various renamings in tests such that common names are used. [\#119](https://github.com/arangodb/kube-arangodb/pull/119)
- Added envvar \(CLEANUPDEPLOYMENTS\) to cleanup failed tests. [\#118](https://github.com/arangodb/kube-arangodb/pull/118)
- Added test that removes PV, PVC & Pod or dbserver. \[ci VERBOSE=1\] \[ci LONG=1\] \[ci TESTOPTIONS="-test.run ^TestResiliencePVDBServer$"\] [\#117](https://github.com/arangodb/kube-arangodb/pull/117)
- Fixed expected value for ENGINE file in init container of dbserver. [\#116](https://github.com/arangodb/kube-arangodb/pull/116)
- Improved liveness detection [\#115](https://github.com/arangodb/kube-arangodb/pull/115)
- Run chaos-monkey in go-routine to avoid blocking the operator [\#114](https://github.com/arangodb/kube-arangodb/pull/114)
- Added examples for exposing metrics to Prometheus [\#113](https://github.com/arangodb/kube-arangodb/pull/113)
- Replace HTTP server with HTTPS server [\#112](https://github.com/arangodb/kube-arangodb/pull/112)
- Disabled colorizing logs [\#111](https://github.com/arangodb/kube-arangodb/pull/111)
- Safe resource watcher [\#110](https://github.com/arangodb/kube-arangodb/pull/110)
- Archive log files [\#109](https://github.com/arangodb/kube-arangodb/pull/109)
- Doc - Follow file name conventions of main docs, move to Tutorials [\#108](https://github.com/arangodb/kube-arangodb/pull/108)
- Quickly fail when deployment no longer exists [\#107](https://github.com/arangodb/kube-arangodb/pull/107)
- BREAKING CHANGE: Renamed all enum values to title case [\#104](https://github.com/arangodb/kube-arangodb/pull/104)
- Changed TLSSpec.TTL to new string based `Duration` type [\#103](https://github.com/arangodb/kube-arangodb/pull/103)
- Added automatic renewal of TLS server certificates [\#102](https://github.com/arangodb/kube-arangodb/pull/102)
- Adding GettingStarted page and structuring docs for website [\#101](https://github.com/arangodb/kube-arangodb/pull/101)
- Added LivenessProbe & Readiness probe [\#100](https://github.com/arangodb/kube-arangodb/pull/100)
- Patch latest version number in README [\#99](https://github.com/arangodb/kube-arangodb/pull/99)
- Adding CHANGELOG.md generation [\#98](https://github.com/arangodb/kube-arangodb/pull/98)
- Adding chaos-monkey for deployments [\#96](https://github.com/arangodb/kube-arangodb/pull/96)
- Check contents of persisted volume when dbserver is restarting [\#95](https://github.com/arangodb/kube-arangodb/pull/95)
- Added helper to prepull arangodb \(enterprise\) image. This allows the normal tests to have decent timeouts while prevent a timeout caused by a long during image pull. [\#94](https://github.com/arangodb/kube-arangodb/pull/94)
- Fixing PV cleanup [\#93](https://github.com/arangodb/kube-arangodb/pull/93)
- Check member failure [\#92](https://github.com/arangodb/kube-arangodb/pull/92)
- Tracking recent pod terminations [\#91](https://github.com/arangodb/kube-arangodb/pull/91)
- Enable LONG on kube-arangodb-long test [\#90](https://github.com/arangodb/kube-arangodb/pull/90)
- Tests/multi deployment [\#89](https://github.com/arangodb/kube-arangodb/pull/89)
- Tests/modes [\#88](https://github.com/arangodb/kube-arangodb/pull/88)
- increase timeout for long running tests [\#87](https://github.com/arangodb/kube-arangodb/pull/87)
- fix rocksdb\_encryption\_test [\#86](https://github.com/arangodb/kube-arangodb/pull/86)
- fix - /api/version will answer on all servers \(not leader only\) [\#85](https://github.com/arangodb/kube-arangodb/pull/85)
- fixes required after merge [\#84](https://github.com/arangodb/kube-arangodb/pull/84)
- Deployment state -\> phase [\#83](https://github.com/arangodb/kube-arangodb/pull/83)
- Added detection on unschedulable pods [\#82](https://github.com/arangodb/kube-arangodb/pull/82)
- AsOwner no longer things the owner refers to a controller. It refers to the ArangoDeployment [\#81](https://github.com/arangodb/kube-arangodb/pull/81)
- Store & compare hash of secrets. [\#80](https://github.com/arangodb/kube-arangodb/pull/80)
- Control jenkins from git commit log. [\#79](https://github.com/arangodb/kube-arangodb/pull/79)
- Fix scale-up [\#78](https://github.com/arangodb/kube-arangodb/pull/78)
- Added terminated-pod cleanup to speed up re-creation of pods. [\#77](https://github.com/arangodb/kube-arangodb/pull/77)
- add upgrade tests [\#76](https://github.com/arangodb/kube-arangodb/pull/76)
- check result of api version call [\#75](https://github.com/arangodb/kube-arangodb/pull/75)
- Also watch changes in PVCs and Services [\#74](https://github.com/arangodb/kube-arangodb/pull/74)
- Feature/test individual pod deletion [\#72](https://github.com/arangodb/kube-arangodb/pull/72)
- Moved low level resource \(pod,pvc,secret,service\) creation & inspection to resources sub-package. [\#71](https://github.com/arangodb/kube-arangodb/pull/71)
- Moved reconciliation code to separate package [\#70](https://github.com/arangodb/kube-arangodb/pull/70)
- Test/different deployments resilient [\#69](https://github.com/arangodb/kube-arangodb/pull/69)
- Store accepted spec [\#68](https://github.com/arangodb/kube-arangodb/pull/68)
- Fixed behavior for scaling UI integration wrt startup of the cluster [\#67](https://github.com/arangodb/kube-arangodb/pull/67)
- Fixed immitable `mode` field. [\#66](https://github.com/arangodb/kube-arangodb/pull/66)
- Integrate with scaling web-UI [\#65](https://github.com/arangodb/kube-arangodb/pull/65)
- add test for different deployments [\#64](https://github.com/arangodb/kube-arangodb/pull/64)
- Fixed validation of tls.caSecretName=None [\#61](https://github.com/arangodb/kube-arangodb/pull/61)
- Feature/add tests for immutable cluster parameters [\#59](https://github.com/arangodb/kube-arangodb/pull/59)
- rename test function [\#58](https://github.com/arangodb/kube-arangodb/pull/58)
- Detecting ImageID & ArangoDB version. [\#57](https://github.com/arangodb/kube-arangodb/pull/57)
- Adds ssl support for scaling test [\#53](https://github.com/arangodb/kube-arangodb/pull/53)
- Rotation support for members. [\#49](https://github.com/arangodb/kube-arangodb/pull/49)
- begin to add tests for `apis/storage/v1alpha` [\#36](https://github.com/arangodb/kube-arangodb/pull/36)

## [0.0.1](https://github.com/arangodb/kube-arangodb/tree/0.0.1) (2018-03-20)
**Merged pull requests:**

- Changed scope of ArangoLocalStorage to Cluster. [\#56](https://github.com/arangodb/kube-arangodb/pull/56)
- External crd creation [\#55](https://github.com/arangodb/kube-arangodb/pull/55)
- Rename default docker image to kube-arangodb [\#54](https://github.com/arangodb/kube-arangodb/pull/54)
- Splitting operator in two parts [\#52](https://github.com/arangodb/kube-arangodb/pull/52)
- Turn on TLS by default [\#51](https://github.com/arangodb/kube-arangodb/pull/51)
- Rename repository to `kube-arangodb` [\#48](https://github.com/arangodb/kube-arangodb/pull/48)
- Use single image tag to prevent polluting the docker hub [\#47](https://github.com/arangodb/kube-arangodb/pull/47)
- Renamed pkg/apis/arangodb to pkg/apis/deployment [\#46](https://github.com/arangodb/kube-arangodb/pull/46)
- Added release code [\#45](https://github.com/arangodb/kube-arangodb/pull/45)
- Cleaning up deployment, avoiding docker overrides [\#44](https://github.com/arangodb/kube-arangodb/pull/44)
- TLS support [\#43](https://github.com/arangodb/kube-arangodb/pull/43)
- Adds "Storage Resource" to user README [\#42](https://github.com/arangodb/kube-arangodb/pull/42)
- Reworked TLS spec [\#41](https://github.com/arangodb/kube-arangodb/pull/41)
- Set sesion affinity for coordinator [\#40](https://github.com/arangodb/kube-arangodb/pull/40)
- Set PublishNotReadyAddresses on coordinator&syncmasters service [\#39](https://github.com/arangodb/kube-arangodb/pull/39)
- Prepare test cluster [\#38](https://github.com/arangodb/kube-arangodb/pull/38)
- Run tests on multiple clusters in parallel [\#37](https://github.com/arangodb/kube-arangodb/pull/37)
- Implemented isDefault behavior of storage class [\#35](https://github.com/arangodb/kube-arangodb/pull/35)
- add some tests for util/k8sutil/erros.go [\#32](https://github.com/arangodb/kube-arangodb/pull/32)
- Adding `ArangoLocalStorage` resource \(wip\) [\#31](https://github.com/arangodb/kube-arangodb/pull/31)
- Added custom resource spec for ArangoDB Storage operator. [\#30](https://github.com/arangodb/kube-arangodb/pull/30)
- Added unit tests for k8s secrets & utility methods [\#28](https://github.com/arangodb/kube-arangodb/pull/28)
- Added unit test for creating affinity [\#27](https://github.com/arangodb/kube-arangodb/pull/27)
- More simple tests [\#26](https://github.com/arangodb/kube-arangodb/pull/26)
- Changed default storage engine to RocksDB [\#24](https://github.com/arangodb/kube-arangodb/pull/24)
- Adding command line tests for arangod commandlines. [\#23](https://github.com/arangodb/kube-arangodb/pull/23)
- UnitTests for plan\_builder [\#22](https://github.com/arangodb/kube-arangodb/pull/22)
- Unit tests for apis/arangodb/v1alpha package [\#21](https://github.com/arangodb/kube-arangodb/pull/21)
- Fix bash error [\#20](https://github.com/arangodb/kube-arangodb/pull/20)
- Renamed Controller to Operator [\#19](https://github.com/arangodb/kube-arangodb/pull/19)
- Cleanup kubernetes after tests [\#18](https://github.com/arangodb/kube-arangodb/pull/18)
- Adding rocksdb encryption key support [\#17](https://github.com/arangodb/kube-arangodb/pull/17)
- Adding test design [\#16](https://github.com/arangodb/kube-arangodb/pull/16)
- avoid sub-shell creation [\#15](https://github.com/arangodb/kube-arangodb/pull/15)
- Adding authentication support [\#14](https://github.com/arangodb/kube-arangodb/pull/14)
- Scaling deployments [\#13](https://github.com/arangodb/kube-arangodb/pull/13)
- Test framework [\#11](https://github.com/arangodb/kube-arangodb/pull/11)
- Change docs to "authentication default on" [\#10](https://github.com/arangodb/kube-arangodb/pull/10)
- Pod monitoring [\#9](https://github.com/arangodb/kube-arangodb/pull/9)
- Pod affinity [\#8](https://github.com/arangodb/kube-arangodb/pull/8)
- Extended storage docs wrt local storage [\#7](https://github.com/arangodb/kube-arangodb/pull/7)
- Adding event support [\#6](https://github.com/arangodb/kube-arangodb/pull/6)
- Added pod probes [\#5](https://github.com/arangodb/kube-arangodb/pull/5)
- Creating pods [\#4](https://github.com/arangodb/kube-arangodb/pull/4)
- Extending spec & status object. Implementing service & pvc creation [\#3](https://github.com/arangodb/kube-arangodb/pull/3)
- Initial API objects & vendoring [\#2](https://github.com/arangodb/kube-arangodb/pull/2)
- Added specification of custom resource [\#1](https://github.com/arangodb/kube-arangodb/pull/1)



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*
