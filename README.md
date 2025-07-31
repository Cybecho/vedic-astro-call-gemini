# 베다 점성술 AI 해석 서비스

Google Gemini AI를 활용한 베다 점성술 차트 해석 서비스입니다.

## 개요

이 서비스는 베다 점성술 차트 JSON 데이터를 입력받아 Google Gemini 2.5 Flash Lite 모델을 통해 시적이고 영혼을 위로하는 해석을 제공합니다.

## 주요 기능

- HTTP POST 요청으로 베다 점성술 차트 데이터 수신
- Google Gemini AI를 통한 전문적이고 시적인 해석 생성
- 구조화된 JSON 응답
- 포괄적인 에러 처리
- Docker 컨테이너 지원

## 기술 스택

- **언어**: Go 1.24
- **AI 모델**: Google Gemini 2.5 Flash Lite
- **포트**: 9494
- **배포**: Docker

## 설치 및 실행

### 로컬 실행

1. 저장소 클론
```bash
git clone https://github.com/Cybecho/vedic-astro-call-gemini.git
cd vedic-astro-call-gemini
```

2. 의존성 설치
```bash
go mod tidy
```

3. 환경 변수 설정
```bash
export GEMINI_API_KEY="your-gemini-api-key"
```

4. 서버 실행
```bash
go run main.go
```

### Docker 실행

1. Docker 이미지 빌드
```bash
docker build -t vedic-astrology-service .
```

2. 컨테이너 실행
```bash
docker run -d -p 9494:9494 -e GEMINI_API_KEY="your-api-key" vedic-astrology-service
```

## API 사용법

### 요청

- **엔드포인트**: `POST /interpret`
- **Content-Type**: `application/json`
- **포트**: 9494

### 요청 형식

```json
{
  "chart": {
    "user": {
      "datetime": "2000-02-07 06:00:00",
      "timezone": "+09:00",
      "longitude": 127.0286,
      "latitude": 37.2635,
      "altitude": 0
    },
    "graha": {
      "Su": { /* 태양 데이터 */ },
      "Mo": { /* 달 데이터 */ },
      // ... 기타 행성 데이터
    },
    // ... 기타 차트 데이터
  },
  "duration_of_response": 0.012,
  "created_at": "2025-07-31 06:58:02"
}
```

### 응답 형식

#### 성공 응답
```json
{
  "interpretation": "해석 결과 텍스트...",
  "success": true
}
```

#### 에러 응답
```json
{
  "interpretation": "",
  "success": false,
  "error": "에러 메시지"
}
```

### 테스트

제공된 샘플 데이터로 테스트:

```bash
curl -X POST http://localhost:9494/interpret \
  -H "Content-Type: application/json" \
  -d @sample-request.txt
```

## 파일 구조

```
├── main.go              # 메인 서버 코드
├── post-prompt.txt      # AI 해석을 위한 프롬프트 템플릿
├── sample-request.txt   # 테스트용 샘플 요청 데이터
├── go.mod              # Go 모듈 설정
├── go.sum              # 의존성 체크섬
├── Dockerfile          # Docker 이미지 설정
├── .dockerignore       # Docker 빌드 제외 파일
└── README.md           # 프로젝트 문서
```

## 보안 고려사항

- Gemini API 키는 환경 변수로 관리
- 30초 요청 타임아웃 설정
- 입력 데이터 검증
- 구조화된 에러 응답

## 라이센스

이 프로젝트는 MIT 라이센스 하에 배포됩니다.
