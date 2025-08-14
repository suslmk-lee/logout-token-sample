# 멀티스테이지 빌드
# 1단계: 빌드 스테이지
FROM golang:1.21-alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app

# Go 모듈 파일 복사
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# CGO 비활성화하고 정적 바이너리 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 2단계: 실행 스테이지
FROM alpine:latest

# 보안 업데이트 및 ca-certificates 설치 (HTTPS 요청용)
RUN apk --no-cache add ca-certificates tzdata

# 작업 디렉토리 설정
WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/main .

# 포트 노출
EXPOSE 3001

# 애플리케이션 실행
CMD ["./main"]