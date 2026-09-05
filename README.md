# WireGuard VPN 및 정책 기반 프록시 시스템 — 프로젝트 가이드
## 1. 프로젝트 개요 및 아키텍처
- 개요: 본 시스템은 클라이언트의 트래픽을 WireGuard VPN 터널을 통해 안전하게 수신하고, 도커(Docker) 환경 내에서 구동되는 고성능 Go 기반 프록시 및 정책 엔진을 통해 트래픽을 선별·분석·검증·제어하는 통합 보안 프록시 인프라입니다.

- 아키텍처 구성 요소:
  * Client: WireGuard 설정 및 키페어를 사용하여 VPN 터널로 트래픽을 전송하는 출발지 장치
  * WireGuard Tunnel & VPN Server: 도커 환경 내에서 암호화된 UDP 터널링 및 종단점(Endpoint) 관리 수행
  * Policy Engine (policy.json): 트래픽 제어 및 필터링 규칙을 참조하여 허용 여부 결정
  * Validation & Proxy (handler.go, parser_http.go): HTTP 요청 처리, CONNECT 터널링 및 TLS 인터셉션 수행
  * HTTP/HTTPS Server: 정제된 트래픽이 도달하는 최종 목적지 웹서버

## 2. 프로젝트 파일 구조
```

proxy_submission/
├── run.sh             # VPN 키 생성 및 도커 환경 통합 실행 스크립트
├── Dockerfile         # 프록시 및 VPN 서버 단일 컨테이너 이미지 빌드 정의
├── docker-compose.yml # 다중 컨테이너 오케스트레이션 및 네트워크 설정
├── main.go            # 프록시 서버 진입점 (Entrypoint)
├── handler.go         # 코어 프록시 핸들러 로직
├── policy.go          # 정책 엔진 로직 (규칙 로드 및 평가)
├── policy.json        # 트래픽 제어 및 필터링 규칙 정의 파일
├── parser_http.go     # HTTP 트래픽 및 헤더 파싱
└── logger.go          # 시스템 이벤트 및 트래픽 로그 추적 관리


```

## 3. 프로젝트 빌드 및 실행 방법
- 사전 요구사항: 호스트 시스템에 Docker, Docker Compose, WireGuard가 설치되어 있어야 합니다.
- 키 생성 및 실행:
  WireGuard 활성화를 위한 공개키와 개인키가 자동으로 생성되며, 자동화된 설정 및 구동 스크립트를 실행합니다.

  ```
  chmod +x run.sh
  ./run.sh
  ```
    chmod +x run.sh 명령은 실행 권한을 run.sh 파일에 추가 합니다 그리고 ./run.sh 는 프록시 및 Docker 환경을 동작시킵니다.
 ## 4. 기능별 테스트 방법 및 결과
 - TLS 트래픽 복호화 (인증서 경고 없는 HTTPS 접속):
     * 테스트 방법: 생성된 인증서를 활용한 커스텀 인터셉션 모듈 (handler.go)을 통해 HTTPS 트래픽을 프록시로 라우팅.
     * 테스트 결과: 웹 브라우저에서 보안 경고 없이 HTTPS 사이트에 성공적으로 접근하는 것을 확인했습니다.

       <img width="707" height="57" alt="image" src="https://github.com/user-attachments/assets/7be38bc0-1a50-47ac-9bcb-3a77b6bfc751" />


       <img width="1572" height="747" alt="image" src="https://github.com/user-attachments/assets/4dca125c-9990-4e26-9fb8-80f26d74c1a5" />

 
 - 프로토콜 지원 (HTTP, HTTPS):
     * 테스트 방법: 안전한 터널과 프록시 파이프라인을 통해 다양한 프로토콜 페이로드를 전송 및 프록시 처리.
     * 테스트 결과: HTTP, MySQL, SSH 트래픽에 대한 성공적인 연결 처리 및 트래픽 포워딩을 검증했습니다. (※ 각 프로토콜 테스트 결과 스크린샷 첨부)
  
  ## 5. 고려한 문제점, 해결 방안 및 개선·확장 계획
  - 고려한 문제점 및 해결 방안:
     * 문제점: 클라이언트의 연결 안정성을 해치지 않으면서 암호화된 HTTPS/TLS 트래픽을 안전하게 인터셉트하고 검증하는 문제.
     * 해결 방안: 동적 CONNECT 메서드 핸들러와 TLS 인터셉션 워크플로우(handler.go)를 구현하고, 이를 policy.json의 규칙 검증 엔진과 연동하여 해결했습니다.

  - 개선 및 확장 계획:
     * 고부하 엔터프라이즈 환경을 위해 parser_http.go의 패킷 파싱 성능 최적화.
     * logger.go를 활용한 실시간 트래픽 모니터링 및 로그 대시보드 확장.
 
