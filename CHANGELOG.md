# HyperCloud-Multi-Operator changelog!!
All notable changes to this project will be documented in this file.

<!-------------------- v5.0.34.5 start -------------------->

## HyperCloud-Multi-Operator_5.0.34.5 (2022. 10. 25. (화) 13:46:36 KST)

### Added
  - [feat] single cluster 생성시 etcd listen-metrics-urls에 master ip 추가 by sjoh0704

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.34.5 end --------------------->

<!-------------------- v5.0.34.4 start -------------------->

## HyperCloud-Multi-Operator_5.0.34.4 (2022. 10. 21. (금) 14:46:21 KST)

### Added

### Changed
  - [mod] argocd manager token의 data 값을 제대로 받아오지 못하는 error fix by sjoh0704

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.34.4 end --------------------->

<!-------------------- v5.0.34.3 start -------------------->

## HyperCloud-Multi-Operator_5.0.34.3 (2022. 10. 12. (수) 10:51:34 KST)

### Added

### Changed
  - [mod] bind-address 0.0.0.0으로 변경 by sjoh0704

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.34.3 end --------------------->

<!-------------------- v5.0.34.2 start -------------------->

## HyperCloud-Multi-Operator_5.0.34.2 (2022. 09. 30. (금) 15:37:17 KST)

### Added

### Changed
  - [mod] single cluster에 hyperauth tls certificate 생성하는 로직 제거 by sjoh0704

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.34.2 end --------------------->

<!-------------------- v5.0.34.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.34.1 (2022. 09. 29. (목) 15:43:23 KST)

### Added
  - [feat] 필요 환경변수가 있는지 check하는 기능 추가 by sjoh0704

### Changed
  - [mod] clustertemplate 내 hyperauth 변수 처리 by sjoh0704

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.34.1 end --------------------->

<!-------------------- v5.0.34.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.34.0 (2022. 09. 21. (수) 16:09:17 KST)

### Added

### Changed
  - [mod] fix errors in delete reconcilation by seung
  - [mod] kubeconfig delete logic 변경 및 cluster registration delete logic 변경 by seung
  - [mod] kiali hyperauth 연동시 default client scope로 생성 by seung
  - [mod] fix bug in get user-id by email api call by seung
  - [mod] hyperauth user api 대신 keycloak user api를 사용하도록 변경 by seung

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.34.0 end --------------------->

<!-------------------- v5.0.33.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.33.1 (2022. 09. 08. (목) 16:37:47 KST)

### Added

### Changed
  - [mod] fix create hyperauth client error/disable DeleteDeprecatedTraefikResources by seung

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.33.1 end --------------------->

<!-------------------- v5.0.33.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.33.0 (2022. 09. 02. (금) 14:52:23 KST)

### Added

### Changed
  - [mod] HyperAuthError type 로직 error 수정 by SISILIA

### Fixed
  - [ims][289299] hyperregistry oidc연동 방식 변경에 따른 코드 수정 by SISILIA

### CRD yaml

### Etc
  - [etc] 불필요한 코드 제거 by SISILIA
  - [etc] 오타 수정 by SISILIA
  - [etc] cluster status 로직 일부 수정 by SISILIA
  - [etc] status 체크 에러 수정 by SISILIA
  - [etc] cluster claim reconcile 로직 error 수정 by SISILIA
  - [etc] status 변경 migration 로직 보강 by SISILIA
  - [etc] application link status migration 로직 추가 by SISILIA
  - [etc] application Link 생성 로직 추가 by SISILIA

<!--------------------- v5.0.33.0 end --------------------->

<!-------------------- v5.0.33.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.33.0 (2022. 09. 02. (금) 14:23:32 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] 오타 수정 by SISILIA
  - [etc] cluster status 로직 일부 수정 by SISILIA
  - [etc] status 체크 에러 수정 by SISILIA
  - [etc] cluster claim reconcile 로직 error 수정 by SISILIA
  - [etc] status 변경 migration 로직 보강 by SISILIA
  - [etc] application link status migration 로직 추가 by SISILIA
  - [etc] application Link 생성 로직 추가 by SISILIA

<!--------------------- v5.0.33.0 end --------------------->

<!-------------------- v5.0.32.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.32.0 (2022. 08. 24. (수) 12:51:49 KST)

### Added
  - [feat] cluster tab status 기획 수정에 따른 코드 수정 by SISILIA
  - [feat] 멀티클러스터 스테이터스 기획수정에 따른 상태 추가 및 로직 수정 by SISILIA
  - [feat] single application을 multi-operator에서 생성해주는 로직 추가 by SISILIA

### Changed
  - [mod] v1beta1을 위해 preserveUnknownFields를 넣는 방식을 Makefile option으로 변경 by SISILIA
  - [mod] log level 적용안된 부분 추가 by seung
  - [mod] cluster claim이 rejected일때는 spec수정이 가능하도록 웹훅로직 수정 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] 불필요한 코드 제거 by SISILIA
  - [etc] cluster role binding spec error 수정 by SISILIA
  - [etc] 주석 작업 by SISILIA
  - [etc] 오타수정 by SISILIA
  - [etc] crd v1beta1 explain을 위한 spec 추가 by SISILIA

<!--------------------- v5.0.32.0 end --------------------->

<!-------------------- v5.0.31.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.31.0 (2022. 08. 12. (금) 10:38:29 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.31.0 end --------------------->

<!-------------------- v5.0.30.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.30.0 (2022. 08. 02. (화) 13:09:21 KST)

### Added
  - [feat] claim 생성시 clustermanager 생성 기능 추가 by seung

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] 오타수정 by seung

<!--------------------- v5.0.30.0 end --------------------->

<!-------------------- v5.0.29.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.29.0 (2022. 08. 02. (화) 13:03:22 KST)

### Added

### Changed
  - [mod] capi template에서 더이상 ingress-nginx 설치하지 않도록 수정 by SISILIA

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.29.0 end --------------------->

<!-------------------- v5.0.28.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.28.0 (2022. 08. 02. (화) 12:29:00 KST)

### Added
  - [feat] 클러스터 삭제전에 argocd application이 모두 삭제되었는지 체크하는 로직 추가 by SISILIA

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.28.0 end --------------------->

<!-------------------- v5.0.27.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.27.0 (2022. 08. 01. (월) 17:46:44 KST)

### Added
  - [feat] jwt-decode-auth에서 사용할 service account token secret 생성 및 관리 로직 추가 by SISILIA
  - [feat] kube-v1.22용 kustomize build가 가능하도록 /config/crd/patches 하위 manifest 수정 by SISILIA
  - [feat] kube-v1.22 에서 호환 가능하도록 crd, webhook manifest를 v1에 맞게 수정 by SISILIA

### Changed
  - [mod] argocd-installer에서 timezone을 env가 아닌 mount로 받을 수 있도록 TZ env 제거 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] patch b5.0.26.12 manifests by SISILIA
  - [etc] b5.0.26.x 버전 패치사항을 최신버전에도 적용될 수 있도록 코드 merge by SISILIA
  - [etc] 불필요한 주석 제거 by SISILIA

<!--------------------- v5.0.27.0 end --------------------->

<!-------------------- v5.0.26.16 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.16 (2022. 07. 29. (금) 12:50:35 KST)

### Added
  - [feat] cluster가 남아있는 경우, cluster claim이나 cluster registration을 삭제하지 못하도록 웹훅추가 by SISILIA

### Changed
  - [mod] gateway service suffix를 고려하여 웹훅 로직 수정 by SISILIA
  - [mod] webhook error 수정 by SISILIA
  - [mod] webhook 로직 에러 수정 by SISILIA
  - [mod] k8s service 이름이 DNS-1035룰을 따르므로 cluster name도 DNS-1123룰에서 DNS-1035룰로 변경 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] webhook error fix by SISILIA
  - [etc] 일부 주석 제거 by SISILIA

<!--------------------- v5.0.26.16 end --------------------->

<!-------------------- v5.0.26.15 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.15 (2022. 07. 27. (수) 11:45:47 KST)

### Added
  - [feat] opensearch-developer, opensearch-guest client role 추가 by SISILIA

### Changed
  - [mod] argocd service account token secret을 찾지 못하는 에러 수정 by SISILIA

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.26.15 end --------------------->

<!-------------------- v5.0.26.14 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.14 (2022. 07. 25. (월) 09:38:00 KST)

### Added
  - [feat] log level설정할 수 있도록 코드 추가 및 yaml 수정 by seung 
### Changed
  - [mod] traefik service가 single cluster에 아직 설치되지 않았을 경우, error가 아닌 info로 로그 레벨 변경 by SISILIA
  - [mod] lb type의 svc를 지우는 로직에서 error 반환 로직 수정 by SISILIA

### Fixed
  - [ims][286609] capa nlb deletion error 해결을 위하여 클러스터 삭제전 lb타입의 svc를 모두 삭제하는 로직 추가 by SISILIA

### CRD yaml

### Etc
  - [etc] 불필요한 에러로그 제거 by SISILIA

<!--------------------- v5.0.26.14 end --------------------->

<!-------------------- v5.0.26.13 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.13 (2022. 07. 06. (수) 16:42:06 KST)

### Added

### Changed
  - [mod] ims-285744 변경사항 roll back by SISILIA

### Fixed
  - [ims][285744] self-signed 인증서를 사용해도 정상 동작하도록 hyperauth, hyperregistry에 http콜을 insecure로 날리도록 변경 by SISILIA

### CRD yaml

### Etc

<!--------------------- v5.0.26.13 end --------------------->

<!-------------------- v5.0.26.12 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.12 (2022. 06. 23. (목) 11:14:42 KST)

### Added
  - [feat] capi가 생성해주는 kubeconfig secret에 대한 관리 로직 추가 by SISILIA
  - [feat] 불필요한 reconcile 수행 및 로그를 없애기 위해 capi로 cluster생성시에 control plane이 ready되지 않으면 clm controller reconcile이 수행되지 않게 로직 추가 by SISILIA

### Changed

### Fixed
  - [ims][285253] calico 버전 고정을 위한 template 수정 by SISILIA

### CRD yaml

### Etc
  - [etc] golang version에 맞게 makefile 수정 by SISILIA

<!--------------------- v5.0.26.12 end --------------------->

<!-------------------- v5.0.26.11 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.11 (2022. 05. 19. (목) 13:46:44 KST)

### Added
  - [feat] 클러스터 등록 삭제시 초대된 member에 대한 cluster role binding을 삭제하는 로직 추가 by sihyunglee823
  - [feat] hyperregistry를 설치 하지 않는 경우에는 oidc 설정 phase를 skip하는 로직 추가 by SISILIA
  - [feat] 클러스터 등록 삭제시 argocd-manager service account token secret 삭제 기능 추가 by SISILIA
  - [feat] spec.clusterName의 글자수를 체크하는 webhook 로직 추가 by SISILIA

### Changed
  - [mod] argocd-manager service account token secret 스펙 에러 수정 by SISILIA
  - [mod] webhook 로직 if문 조건 중복 에러 수정 by SISILIA
  - [mod] ArgoCD 연동을 위한 secret을 생성할 때, admin정보가 아닌 service account정보를 사용하도록 변경 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] 주석 수정 by sihyunglee823
  - [etc] 주석 추가 by sihyunglee823
  - [etc] 주석 수정 by sihyunglee823
  - [etc] 주석추가 및 오타수정 by sihyunglee823
  - [etc] 주석추가 및 오타수정 by sihyunglee823
  - [etc] 주석 추가 by SISILIA
  - [etc] log 에러 수정, 누락된 로그 추가 by SISILIA
  - [etc] json marshal 에러 발생시 방어로직 추가 by SISILIA
  - [etc] 웹훅 로직에 DNS-1123룰에 대한 설명 및 웹훅로직 설명 주석 추가 by SISILIA
  - [etc] 주석 보강 by SISILIA
  - [etc] fix changelog.md typo by SISILIA

<!--------------------- v5.0.26.11 end --------------------->

<!-------------------- v5.0.26.10 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.10 (2022. 04. 26. (화) 11:17:16 KST)

### Added

### Changed
  - [mod] gateway tls secret 환경변수 제거 by SISILIA

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.26.10 end --------------------->

<!-------------------- v5.0.26.9 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.9 (2022. 04. 26. (화) 10:50:21 KST)

### Added

### Changed
  - [mod] opensearch용 시크릿 배포 로직 제거 by SISILIA

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.26.9 end --------------------->

<!-------------------- v5.0.26.8 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.8 (2022. 04. 21. (목) 17:11:41 KST)

### Added
  - [feat] opensearch용 hyperauth client 생성 로직 추가 및 tls secret deploy 로직 추가 by SISILIA

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.26.8 end --------------------->

<!-------------------- v5.0.26.7 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.7 (2022. 04. 15. (금) 16:04:27 KST)

### Added
  - [feat] hyperregistry oidc 연동 기능 추가 by SISILIA
  - [feat] hyperauth client secret env 추가 by SISILIA
  - [feat] single cluster kiali client 세부 설정 추가 by SISILIA
  - [feat] Single cluster를 위한 kibana client생성시, role생성, user mapping도 해주도록 로직 추가 by SISILIA
  - [feat] hyperauth client 생성 기능 추가 by SISILIA
  - [feat] kibana를 위한 ingress path 추가 by SISILIA

### Changed
  - [mod] hyperauth 주소를 사용자가 설정한 subdomain으로 처리할 수 있도록 수정 by SISILIA
  - [mod] hyperauth caller package 분리 및 refactoring by SISILIA
  - [mod] webhook 이름 변경, webhook에 clm, clr의 spec.clusterName validation 로직 추가 by SISILIA
  - [mod] 불필요한 서비스 생성하지 않도록 코드 수정 by SISILIA
  - [mod] memory limit 100Mi로 상향 by SISILIA
  - [mod] endpoint가 ip일때도 ep 객체를 생성하지 않고 service의 externalName에 넣도록 변경 by SISILIA

### Fixed

### CRD yaml

### Etc
  - [etc] kustomize manifest.yaml 에러 수정 by SISILIA
  - [etc] code refactoring by SISILIA
  - [etc] gitignore update by SISILIA
  - [etc] code refactoring, warning 발생시키는 코드 수정 by SISILIA
  - [etc] code refactoring by SISILIA
  - [etc] argo status update 버그 수정 by SISILIA

<!--------------------- v5.0.26.7 end --------------------->

<!-------------------- v5.0.26.6 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.6 (2022. 03. 17. (목) 17:00:04 KST)

### Added

### Changed
  - [mod] argocd cluster role binding ns 에러 수정 by SISILIA

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.26.6 end --------------------->

<!-------------------- v5.0.26.5 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.5 (2022. 02. 25. (금) 15:22:37 KST)

### Added
  - [feat] single cluster를 위한 리소스(Ingress, Certificate etc.)가 삭제되면 재생성 되는 로직 추가 by SISILIA

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] cluster 삭제시 secret 삭제되지 않는 버그 수정 by SISILIA
  - [etc] code refactoring by SISILIA

<!--------------------- v5.0.26.5 end --------------------->

<!-------------------- v5.0.26.4 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.4 (2022. 02. 15. (화) 10:36:53 KST)

### Added
  - [feat] prometheus를 위한 service 및 endponit 생성 로직 추가 by SISILIA
  - [feat] argocd 클러스터 등록기능 추가 by SISILIA
  - [feat] serviceaccount 변경으로 인한 capi를 위한 cluster role 추가 by SISILIA

### Changed
  - [mod] service 및 endpoint 생성 로직분기 기준을 provider가 아닌 endpoint가 ip address / dns 분기로 변경 by SISILIA

### Fixed
  - [ims][276626] webhookconfiguration admissionreviewversions bug 수정 by SISILIA

### CRD yaml

### Etc
  - [etc] code refactoring by SISILIA
  - [etc] remove temp files by SISILIA
  - [etc] fix merge conflict by SISILIA
  - [etc] merge shinhan2 to main by SISILIA
  - [etc] update git ignore by SISILIA
  - [etc] 불필요한 주석 제거 by SISILIA
  - [etc] CRD version v1beta1으로 변경 by SISILIA
  - [etc] webhook configuration v1 sideEffects 버그 수정 by SISILIA
  - [etc] webhook configuration v1 버그 수정 by SISILIA
  - [etc] 오타 수정 by SISILIA

<!--------------------- v5.0.26.4 end --------------------->

<!-------------------- v5.0.26.2 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.2 (2022. 02. 04. (금) 21:05:50 KST)

### Added
  - [feat] role에 leases.coordination.k8s.io에 대한 create, delete, patch verb추가 by SISILIA
  - [etc] 클러스터 등록시, ArgoCD resource 생성 초안 추가, service annotation error수정, 코드리팩터링 by SISILIA
  - [etc] v5.0.26.1에서 hypercloud-multi-operator-manager-role에 leases.coordination.k8s.io에 대한 create, delete, patch verb추가 by SISILIA

<!--------------------- v5.0.26.2 end --------------------->

<!-------------------- v5.0.26.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.1 (2022. 01. 27. (목) 19:44:33 KST)

### Added
  - [feat] 하이퍼클라우드 커스텀 도메인 env 추가 by SISILIA
  - [etc] rbac 리소스 권한 추가 by SISILIA
  - [etc] rbac 리소스 권한 추가 by SISILIA

<!--------------------- v5.0.26.1 end --------------------->

<!-------------------- v5.0.26.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.26.0 (2022. 01. 21. (금) 13:40:34 KST)

### Added
  - [feat] traefik resource관리를 위한 cluster role 추가 by SISILIA
  - [feat] traefik resource 생성 기능 추가 by SISILIA
  - [feat] traefik resource 생성 기능 추가 by SISILIA

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.26.0 end --------------------->

<!-------------------- v5.0.25.19 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.19 (2021. 12. 20. (월) 18:04:06 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] CRD version v1beta1으로 변경 by SISILIA

<!--------------------- v5.0.25.19 end --------------------->

<!-------------------- v5.0.25.18 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.18 (2021. 12. 15. (수) 11:48:10 KST)

### Added
  - [feat] serviceaccount 변경으로 인한 capi를 위한 cluster role 추가 by SISILIA

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] webhook configuration v1 sideEffects 버그 수정 by SISILIA

<!--------------------- v5.0.25.18 end --------------------->

<!-------------------- v5.0.25.17 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.17 (2021. 12. 14. (화) 13:11:45 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] webhook configuration v1 버그 수정 by SISILIA

<!--------------------- v5.0.25.17 end --------------------->

<!-------------------- v5.0.25.16 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.16 (2021. 12. 13. (월) 18:06:01 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] webhook admission review version list에 v1beta1 추가 by SISILIA
  - [etc] cluster manager status 버그 수정 by SISILIA
  - [etc] 오타 수정 by SISILIA
  - [etc] CRD v1beta1으로 변경, 코드 refactoring by SISILIA
  - [etc] cluster registration으로 만들어진 clm삭제시 리소스 삭제가 안되던 버그 수정 by SISILIA
  - [etc] service account token secret 오타 수정 by SISILIA
  - [etc] kustomization file 버그 수정 by SISILIA
  - [etc] webhook configuration v1 버그 수정 by SISILIA

<!--------------------- v5.0.26.0 end --------------------->

<!-------------------- v5.0.25.15 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.15 (2021. 12. 09. (목) 14:13:34 KST)

### Added
  - [feat] service account를 위한 token secret추가 및 volumeMount, volume 설정 추가 by SISILIA

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.25.15 end --------------------->

<!-------------------- v5.0.25.14 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.14 (2021. 12. 08. (수) 16:18:02 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] service account namespace 오타 수정 by SISILIA

<!--------------------- v5.0.25.14 end --------------------->

<!-------------------- v5.0.25.13 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.13 (2021. 12. 08. (수) 15:24:21 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] service account 오타 수정 by SISILIA

<!--------------------- v5.0.25.13 end --------------------->

<!-------------------- v5.0.25.12 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.12 (2021. 12. 08. (수) 15:13:37 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] service account, admission v1 오타 수정 by SISILIA

<!--------------------- v5.0.25.12 end --------------------->

<!-------------------- v5.0.25.11 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.11 (2021. 12. 08. (수) 14:14:35 KST)

### Added
  - [feat] service account 추가 by SISILIA

### Changed

### Fixed

### CRD yaml
  - [crd] k8s 1.22+에서 호환을 위하여 webhook, certificate등의 apiVersion을 v1beta1에서 v1으로 변경 by soohwan kim
  - [crd] k8s 1.22+에서 호환을 위하여 CRD version을 v1beta1에서 v1으로 변경 by soohwan kim

### Etc
  - [etc] vSphere cluster claim spec description 상세 추가 by SISILIA

<!--------------------- v5.0.25.11 end --------------------->

<!-------------------- v5.0.25.10 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.10 (2021. 10. 29. (금) 14:41:55 KST)

### Added

### Changed

### Fixed
  - [ims][273051] clusterclaim CRD description 오류 수정 by soohwan kim
  - [ims][269425] 1670741번 액션 관련 버그 수정 및 clusterregistration 상태 추가 by soohwan kim

### CRD yaml

### Etc

<!--------------------- v5.0.25.10 end --------------------->

<!-------------------- v5.0.25.9 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.9 (2021. 10. 21. (목) 17:29:04 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.25.9 end --------------------->

<!-------------------- v5.0.25.8 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.8 (2021. 10. 13. (수) 15:36:07 KST)

### Added

### Changed

### Fixed
  - [ims][271798] service instance 이름 중복을 피하기 위해 random string을 suffix로 붙여주는 기능 추가 by soohwan kim

### CRD yaml

### Etc
  - [etc] bug fix by soohwan kim
  - [etc] bug fix by soohwan kim

<!--------------------- v5.0.25.8 end --------------------->

<!-------------------- v5.0.25.7 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.7 (2021. 10. 12. (화) 16:49:14 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] cluster manager CRD spec 버그 수정 by soohwan kim

<!--------------------- v5.0.25.7 end --------------------->

<!-------------------- v5.0.25.6 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.6 (2021. 10. 12. (화) 16:44:56 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] CRD spec변경 빌드 by soohwan kim

<!--------------------- v5.0.25.6 end --------------------->

<!-------------------- v5.0.25.5 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.5 (2021. 10. 12. (화) 15:41:51 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.25.5 end --------------------->

<!-------------------- v5.0.25.4 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.4 (2021. 10. 07. (목) 15:44:21 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] cluster manager CRD spec 버그 수정 by soohwan kim

<!--------------------- v5.0.25.4 end --------------------->

<!-------------------- v5.0.25.3 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.3 (2021. 10. 06. (수) 11:02:59 KST)

### Added
  - [feat] 등록하고자 하는 cluster validation 기능 추가 by soohwan kim

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] CRD spec변경 빌드 by soohwan kim
  - [etc] CRD required spec 수정 by soohwan kim
  - [etc] push script 분리 by soohwan kim
  - [etc] CAPI template short description 추가 by soohwan kim

<!--------------------- v5.0.25.3 end --------------------->

<!-------------------- v5.0.25.2 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.2 (2021. 09. 10. (금) 17:17:27 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] cluster template update(tsb 0.2.0 호환) by soohwan kim
  - [etc] push.sh 버전 관리 버그 픽스 by soohwan kim
  - [etc] manager에 Seoul region time zone configuration 추가 by soohwan kim

<!--------------------- v5.0.25.2 end --------------------->

<!-------------------- v5.0.25.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.1 (2021. 09. 07. (화) 13:25:15 KST)

### Added

### Changed

### Fixed
  - [ims][269425] vsphere cluster claim bug fix by soohwan kim
  - [ims][268314] kubeconfig file업로드를 위한 clusterregistration.spec.kubeconfig.format: data-url 추가 by soohwan kim

### CRD yaml

### Etc
  - [etc] push.sh script update, crd update by soohwan kim
  - [etc] spec.kubeconfig.format: data-url 추가 by GitHub

<!--------------------- v5.0.25.1 end --------------------->

<!-------------------- v5.0.25.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.25.0 (2021. 08. 27. (금) 18:55:57 KST)

### Added
  - [Feat] Cluster Claim Spec하위절 구조 변경 by soohwan kim
  - [feat] vsphere cluster claim기능 추가 by soohwan kim

### Changed
  - [Mod] cluster claim, cluster manager 일부스펙 수정 by soohwan kim

### Fixed
  - [ims][268314] 클러스터 등록시 base64인코딩된 kubeconfig파일을 받아서 secret생성 하도록 로직 수정 by soohwan kim

### CRD yaml

### Etc
  - [etc] 테스트 레포 태그 제거 by soohwan kim
  - [etc] commit_rule.md 추가 by soohwan kim
  - [Etc] capi template file추가 by soohwan kim
  - [Etc] cluster claim property name 에러 수정, 불필요한 로깅 제거 by soohwan kim
  - [Etc] 스펙명 오타 수정 by soohwan kim
  - [etc] cluster manager spec에 provider 복구 by soohwan kim
  - [Etc] 코드 indent mismatch 수정, Makefile 주석 수정 by soohwan kim

<!--------------------- v5.0.25.0 end --------------------->

<!-------------------- v5.0.24.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.24.0 (2021. 08. 19. (목) 17:02:50 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.24.0 end --------------------->

<!-------------------- v5.0.23.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.23.0 (2021. 08. 12. (목) 13:05:48 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.23.0 end --------------------->

<!-------------------- v5.0.22.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.22.0 (2021. 08. 05. (목) 15:16:23 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.22.0 end --------------------->

<!-------------------- v5.0.21.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.21.0 (2021. 07. 29. (목) 15:33:04 KST)

### Added
  - [feat] cluster template parameter에 namespace가 추가되어 operator에 처리로직 추가 by soohwan kim

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] image 이름 정책 변경, operator-sdk환경변수 추가 by soohwan kim

<!--------------------- v5.0.21.0 end --------------------->

<!-------------------- v5.0.20.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.20.0 (2021. 07. 22. (목) 14:19:23 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.20.0 end --------------------->

<!-------------------- v5.0.19.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.19.0 (2021. 07. 15. (목) 17:21:19 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] 불필요한 컨틀로러 제거 by chosangwon93

<!--------------------- v5.0.19.0 end --------------------->

<!-------------------- v5.0.18.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.18.1 (2021. 07. 12. (월) 14:16:11 KST)

### Added

### Changed
  - [mod] proxy rule 생성 시 URL 마지막에 / 붙이도록 추가 by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] 코드 정리 by chosangwon93

<!--------------------- v5.0.18.1 end --------------------->

<!-------------------- v5.0.18.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.18.0 (2021. 07. 09. (금) 09:38:11 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.18.0 end --------------------->

<!-------------------- v5.0.17.2 start -------------------->

## HyperCloud-Multi-Operator_5.0.17.2 (2021. 07. 06. (화) 20:43:15 KST)

### Added

### Changed
  - [mod] clustermanager 삭제시 바로 kubefed unjoin하도록 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.17.2 end --------------------->

<!-------------------- v5.0.17.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.17.1 (2021. 07. 05. (월) 15:28:45 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.17.1 end --------------------->

<!-------------------- v5.0.17.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.17.0 (2021. 07. 01. (목) 17:46:14 KST)

### Added
  - [feat] 클러스터 등록기능 추가 by chosangwon93

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.17.0 end --------------------->

<!-------------------- v5.0.16.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.16.0 (2021. 06. 25. (금) 11:12:03 KST)

### Added
  - [feat] HPCD를 통해 생성되지 않은 클러스터를 등록하는 기능 추가 by chosangwon93

### Changed
  - [mod] clustermanager가 존재할 때만 배포된 리소스를 삭제할 수 있도록 수정 by chosangwon93
  - [mod] 등록된 클러스터를 삭제할 때 배포된 role/binding을 제거하도록 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.16.0 end --------------------->

<!-------------------- v5.0.15.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.15.0 (2021. 06. 17. (목) 15:10:50 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.15.0 end --------------------->

<!-------------------- v5.0.14.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.14.0 (2021. 06. 10. (목) 20:22:22 KST)

### Added

### Changed
  - [mod] 클러스터가 삭제되었을 때 db에 직접 접근해서 cluster 정보를 지우지 않고 hypercloud를 호출하도록 수정' by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] .gitignore 수정 by chosangwon93

<!--------------------- v5.0.14.0 end --------------------->

<!-------------------- v5.0.13.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.13.1 (2021. 06. 03. (목) 18:56:06 KST)

### Added

### Changed
  - [mod] 불필요한 delete 이벤트에 대한 webhook 로직 삭제 by chosangwon93

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.13.1 end --------------------->

<!-------------------- v5.0.13.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.13.0 (2021. 06. 03. (목) 17:23:20 KST)

### Added

### Changed

### Fixed

### CRD yaml

### Etc

<!--------------------- v5.0.13.0 end --------------------->

<!-------------------- v5.0.12.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.12.0 (2021. 05. 27. (목) 15:05:09 KST)

### Added

### Changed
  - [mod] 네임스페이스 내에서 클러스터 이름이 중복되지 않게 변경되면서 ClusterManager에서 불필요한 필드 제거 및 로직 수정 by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] add key-mapping files by chosangwon93

<!--------------------- v5.0.12.0 end --------------------->

<!-------------------- v5.0.11.4 start -------------------->

## HyperCloud-Multi-Operator_5.0.11.4 (Tue May 25 05:10:36 KST 2021)

### Added

### Changed
  - [mod] 클러스터 클레임 초기 생성 시 업데이트가 안되는 문제 해결 by chosangwon93
  - [mod] 하이퍼 클라우드 초기 설치 시 마스터 클러스터에 대한 프록시 설정이 되지 않는 문제 해결 by chosangwon93
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.11.4 end --------------------->

<!-------------------- v5.0.11.2 start -------------------->

## HyperCloud-Multi-Operator_5.0.11.2 (Tue May 25 03:00:09 KST 2021)

### Added

### Changed
  - [mod] 하이퍼 클라우드 초기 설치 시 마스터 클러스터에 대한 프록시 설정이 되지 않는 문제 해결 by chosangwon93
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.11.2 end --------------------->
-e 
<!-------------------- v5.0.11.1 start -------------------->
-e 
## HyperCloud-Multi-Operator_5.0.11.1 (2021. 05. 21. (금) 17:31:10 KST)
-e 
### Added
-e 
### Changed
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93
-e 
### Fixed
-e 
### CRD yaml
-e 
### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93
-e 
<!--------------------- v5.0.11.1 end --------------------->

<!-------------------- v5.0.11.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.11.0 (Thu May 20 08:21:10 KST 2021)

### Added

### Changed
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.11.0 end --------------------->

<!-------------------- v5.0.10.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.10.0 (Thu May 13 08:41:47 KST 2021)

### Added

### Changed
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.10.0 end --------------------->

<!-------------------- v5.0.9.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.9.0 (Thu May  6 09:36:33 KST 2021)

### Added

### Changed
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.9.0 end --------------------->

<!-------------------- v5.0.8.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.8.0 (Fri Apr 30 08:58:34 KST 2021)

### Added

### Changed
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.8.0 end --------------------->

<!-------------------- v5.0.7.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.7.0 (Thu Apr 22 10:32:21 KST 2021)

### Added

### Changed
  - [mod] delete aws elb by chosangwon93
  - [mod] add resource req, limit by chosangwon93
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.7.0 end --------------------->

<!-------------------- v5.0.6.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.6.0 (Thu Apr  8 09:41:36 KST 2021)

### Added

### Changed
  - [mod] scope by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] delete duplicated manifests directory by chosangwon93
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.6.0 end --------------------->

<!-------------------- v5.0.3.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.3.1 (Mon Apr  5 02:46:02 KST 2021)

### Added

### Changed
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.3.1 end --------------------->

<!-------------------- v5.0.5.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.5.0 (Thu Apr  1 08:55:44 KST 2021)

### Added

### Changed
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.5.0 end --------------------->

<!-------------------- v5.0.4.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.4.0 (Thu Mar 25 09:30:22 KST 2021)

### Added

### Changed
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.4.0 end --------------------->

<!-------------------- v5.0.3.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.3.0 (Thu Mar 18 17:54:55 KST 2021)

### Added

### Changed
  - [mod] remove status.member by chosangwon93
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.3.0 end --------------------->

<!-------------------- v5.0.2.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.2.0 (Thu Mar 11 19:33:34 KST 2021)

### Added

### Changed
  - [mod] remove status.member by chosangwon93
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.2.0 end --------------------->

<!-------------------- v5.0.1.0 start -------------------->

## HyperCloud-Multi-Operator_5.0.1.0 (Mon Mar  8 15:54:52 KST 2021)

### Added

### Changed
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.1.0 end --------------------->

<!-------------------- v5.0.0.43 start -------------------->

## HyperCloud-Multi-Operator_5.0.0.43 (Mon Mar  8 15:40:09 KST 2021)

### Added

### Changed
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.0.43 end --------------------->

<!-------------------- v5.0.0.42 start -------------------->

## HyperCloud-Multi-Operator_5.0.0.42 (Mon Mar  8 15:33:38 KST 2021)

### Added

### Changed
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.0.42 end --------------------->

<!-------------------- v5.0.0.40 start -------------------->

## HyperCloud-Multi-Operator_5.0.0.40 (Mon Mar  8 15:25:58 KST 2021)

### Added

### Changed
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.0.40 end --------------------->

<!-------------------- v5.0.0.2 start -------------------->

## HyperCloud-Multi-Operator_5.0.0.2 (Tue Mar  2 13:29:05 KST 2021)

### Added

### Changed
  - [mod] add serviceinstance scheme by chosangwon93

### Fixed

### CRD yaml

### Etc
  - [etc] change webhook-server-cert name by chosangwon93
  - [etc] pkg update by chosangwon93
  - [etc] init by chosangwon93

<!--------------------- v5.0.0.2 end --------------------->

<!-------------------- v5.0.0.1 start -------------------->

## HyperCloud-Multi-Operator_5.0.0.1 (Thu Feb 25 13:15:30 KST 2021)

### Added

### Changed

### Fixed

### CRD yaml

### Etc
  - [etc] init by chosangwon93

<!--------------------- v5.0.0.1 end --------------------->
