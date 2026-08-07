# collector/Dockerfile

# Aşama 1: Derleme
FROM golang:1.26-alpine AS builder

# SQLite (C kodları) için gereken C derleyicisini kuruyoruz
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY collector/ ./collector/
WORKDIR /app/collector

# CGO_ENABLED=1 yaparak SQLite köprüsünü aktif ediyoruz
RUN CGO_ENABLED=1 GOOS=linux go build -o collector-bin .

# Aşama 2: Çalıştırma (Hafif Alpine imajı)
FROM alpine:latest  
# tzdata (saat dilimi) paketini de ekliyoruz, loglar için faydalıdır
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/collector/collector-bin .
# Eğer hata verirse config.yaml'ı bulamadığı içindir, bu satırın aktif olduğundan emin ol
COPY collector/config.yaml . 

EXPOSE 8081

CMD ["./collector-bin"]