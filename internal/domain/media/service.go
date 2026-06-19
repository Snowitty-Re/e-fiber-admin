package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"github.com/Snowitty-Re/e-fiber-admin/internal/config"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/media"
	pkgerr "github.com/Snowitty-Re/e-fiber-admin/internal/pkg/errors"
)

type Service struct {
	entClient  *ent.Client
	minioClient *minio.Client
	bucket     string
	useSSL     bool
	endpoint   string
}

func NewService(entClient *ent.Client, minioClient *minio.Client, cfg config.MinIOConfig) *Service {
	return &Service{
		entClient:  entClient,
		minioClient: minioClient,
		bucket:     cfg.Bucket,
		useSSL:     cfg.UseSSL,
		endpoint:   cfg.Endpoint,
	}
}

type UploadResult struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	URL       string `json:"url"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Kind      string `json:"kind"`
}

func (s *Service) Upload(ctx context.Context, fh *multipart.FileHeader) (*UploadResult, error) {
	src, err := fh.Open()
	if err != nil {
		return nil, pkgerr.ErrBadRequest.WithCause(err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	mimeType := fh.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mimeByExt(ext)
	}
	kind := classifyKind(mimeType)

	datePath := time.Now().UTC().Format("2006/01/02")
	key := fmt.Sprintf("media/%s/%s%s", datePath, uuid.NewString(), ext)

	buf := &bytes.Buffer{}
	size, err := io.Copy(buf, src)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	_, err = s.minioClient.PutObject(ctx, s.bucket, key, buf, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload to s3: %w", err)
	}

	url := s.publicURL(key)
	m, err := s.entClient.Media.Create().
		SetKey(key).
		SetURL(url).
		SetMimeType(mimeType).
		SetSizeBytes(size).
		SetKind(media.Kind(kind)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("save media record: %w", err)
	}

	return &UploadResult{
		ID: m.ID, Key: key, URL: url, MimeType: mimeType, SizeBytes: size, Kind: kind,
	}, nil
}

func (s *Service) List(ctx context.Context, kind string, limit, offset int) ([]*ent.Media, int, error) {
	q := s.entClient.Media.Query()
	if kind != "" {
		q = q.Where(media.KindEQ(media.Kind(kind)))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := q.Order(ent.Desc(media.FieldID)).Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) Get(ctx context.Context, id int) (*ent.Media, error) {
	m, err := s.entClient.Media.Query().Where(media.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerr.ErrNotFound
		}
		return nil, fmt.Errorf("query media: %w", err)
	}
	return m, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	m, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	_ = s.minioClient.RemoveObject(ctx, s.bucket, m.Key, minio.RemoveObjectOptions{})
	return s.entClient.Media.DeleteOneID(id).Exec(ctx)
}

func (s *Service) publicURL(key string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, key)
}

func mimeByExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

func classifyKind(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	default:
		return "document"
	}
}
