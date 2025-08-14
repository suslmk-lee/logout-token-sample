# 멀티스테이지 빌드
# 1단계: 빌드 스테이지
FROM node:18-alpine AS builder

# 작업 디렉토리 설정
WORKDIR /app

# package.json과 package-lock.json 복사
COPY package*.json ./

# 의존성 설치 (빌드를 위해 모든 의존성 필요)
RUN npm install

# 소스 코드 복사
COPY . .

# React 앱 빌드
RUN npm run build

# 2단계: 실행 스테이지 (Nginx)
FROM nginx:alpine

# nginx 사용자로 실행할 수 있도록 디렉토리 권한 설정
RUN mkdir -p /var/cache/nginx/client_temp \
    /var/cache/nginx/proxy_temp \
    /var/cache/nginx/fastcgi_temp \
    /var/cache/nginx/uwsgi_temp \
    /var/cache/nginx/scgi_temp \
    && chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && chown -R nginx:nginx /etc/nginx/conf.d \
    && touch /var/run/nginx.pid \
    && chown -R nginx:nginx /var/run/nginx.pid \
    && chown -R nginx:nginx /usr/share/nginx/html

# Nginx 설정 파일 복사
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 빌드된 React 앱을 Nginx로 복사
COPY --from=builder /app/build /usr/share/nginx/html

# nginx 사용자로 소유권 변경
RUN chown -R nginx:nginx /usr/share/nginx/html

# nginx 사용자로 실행
USER nginx

# 포트 노출
EXPOSE 8080

# Nginx 시작
CMD ["nginx", "-g", "daemon off;"]