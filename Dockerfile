# Production image for Render. LibreOffice is included so DOCX/XLSX/PPTX can be
# converted to a temporary PDF for the existing document preview feature.
FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sotaysinhvien .

FROM debian:bookworm-slim
ENV DEBIAN_FRONTEND=noninteractive \
    PORT=10000 \
    APP_ENV=production \
    UPLOAD_DIR=/app/uploads
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      libreoffice-writer libreoffice-calc libreoffice-impress \
      poppler-utils fonts-dejavu-core fonts-liberation \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --uid 10001 appuser \
    && mkdir -p /app/uploads \
    && chown -R appuser:appuser /app /home/appuser
WORKDIR /app
COPY --from=build /out/sotaysinhvien /app/sotaysinhvien
COPY --chown=appuser:appuser config /app/config
USER appuser
EXPOSE 10000
CMD ["/app/sotaysinhvien"]
