# === Stage 1: Build frontend ===
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install --legacy-peer-deps
COPY frontend/ .
RUN npm run build

# === Stage 2: Build backend ===
FROM golang:1.25-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags "-X wa-assistant/backend/license.DevMode=true" -o /wa-assistant ./backend

# === Stage 3: Production ===
FROM nginx:alpine
RUN apk add --no-cache ca-certificates

# Copy backend binary
COPY --from=backend /wa-assistant /app/wa-assistant

# Copy frontend build
COPY --from=frontend /app/frontend/dist /usr/share/nginx/html

# Nginx config: serve frontend + proxy /api to Go backend
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Startup script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 3030
CMD ["/docker-entrypoint.sh"]
