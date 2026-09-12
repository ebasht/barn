package barnmcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ebash/barn/backend/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInvalidSettings = errors.New("invalid MCP settings")

type Settings struct {
	InstanceID    string     `json:"instance_id"`
	InstanceName  string     `json:"instance_name"`
	KeyConfigured bool       `json:"key_configured"`
	AllowWrites   bool       `json:"allow_writes"`
	KeyCreatedAt  *time.Time `json:"key_created_at"`
}

type SettingsService struct {
	db              db.DBTX
	initialToken    string
	initialWrites   bool
	initialInstance InstanceInfo
}

func NewSettingsService(store db.DBTX, token string, writes bool, instance InstanceInfo) *SettingsService {
	return &SettingsService{store, token, writes, instance}
}

// Ensure imports legacy environment settings only when no database row exists.
func (s *SettingsService) Ensure(ctx context.Context) error {
	var hash []byte
	var created *time.Time
	if s.initialToken != "" {
		sum := sha256.Sum256([]byte(s.initialToken))
		hash = sum[:]
		now := time.Now().UTC()
		created = &now
	}
	id := s.initialInstance.ID
	// Default Docker hostnames are not stable identities; persist a generated ID.
	if id == "" {
		id = uuid.NewString()
	}
	name := strings.TrimSpace(s.initialInstance.Name)
	if name == "" {
		name = "Barn"
	}
	_, err := s.db.Exec(ctx, `INSERT INTO mcp_settings(id,instance_id,instance_name,token_hash,allow_writes,key_created_at) VALUES(1,$1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING`, id, name, hash, s.initialWrites, created)
	return err
}

func (s *SettingsService) load(ctx context.Context) (Settings, []byte, error) {
	var settings Settings
	var hash []byte
	err := s.db.QueryRow(ctx, `SELECT instance_id,instance_name,token_hash,allow_writes,key_created_at FROM mcp_settings WHERE id=1`).Scan(&settings.InstanceID, &settings.InstanceName, &hash, &settings.AllowWrites, &settings.KeyCreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		if e := s.Ensure(ctx); e != nil {
			return settings, nil, e
		}
		return s.load(ctx)
	}
	settings.KeyConfigured = len(hash) > 0
	return settings, hash, err
}

func (s *SettingsService) Get(ctx context.Context) (Settings, error) {
	v, _, err := s.load(ctx)
	return v, err
}
func (s *SettingsService) Access(ctx context.Context) (Access, error) {
	v, hash, err := s.load(ctx)
	return Access{Instance: InstanceInfo{ID: v.InstanceID, Name: v.InstanceName, WritesEnabled: v.AllowWrites}, TokenHash: hash}, err
}

func (s *SettingsService) Update(ctx context.Context, name string, writes bool) (Settings, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return Settings{}, fmt.Errorf("%w: panel name must contain 1–200 bytes", ErrInvalidSettings)
	}
	if _, err := s.Get(ctx); err != nil {
		return Settings{}, err
	}
	_, err := s.db.Exec(ctx, `UPDATE mcp_settings SET instance_name=$1,allow_writes=$2,updated_at=now() WHERE id=1`, name, writes)
	if err != nil {
		return Settings{}, err
	}
	return s.Get(ctx)
}

func (s *SettingsService) Generate(ctx context.Context) (string, error) {
	if _, err := s.Get(ctx); err != nil {
		return "", err
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := "barn_mcp_" + hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.Exec(ctx, `UPDATE mcp_settings SET token_hash=$1,key_created_at=now(),updated_at=now() WHERE id=1`, hash[:])
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *SettingsService) Revoke(ctx context.Context) error {
	if _, err := s.Get(ctx); err != nil {
		return err
	}
	_, err := s.db.Exec(ctx, `UPDATE mcp_settings SET token_hash=NULL,key_created_at=NULL,updated_at=now() WHERE id=1`)
	return err
}
