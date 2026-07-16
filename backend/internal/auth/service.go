package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)
import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Sender interface {
	SendCode(context.Context, string, string) error
}
type Metrics interface {
	ObserveAuth(operation, outcome string)
}
type Service struct {
	repo                  Repository
	redis                 *redis.Client
	sender                Sender
	secret                []byte
	accessTTL, refreshTTL time.Duration
	metrics               Metrics
}
type challenge struct {
	Phone    string `json:"phone"`
	DeviceID string `json:"deviceId"`
	CodeMAC  string `json:"codeMac"`
	Attempts int    `json:"attempts"`
}
type Tokens struct {
	AccessToken      string    `json:"accessToken"`
	RefreshToken     string    `json:"refreshToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

var ErrRateLimited = errors.New("rate limited")
var ErrInvalidChallenge = errors.New("invalid challenge")
var ErrInvalidRequest = errors.New("invalid auth request")
var ErrVerificationUnavailable = errors.New("verification unavailable")
var ErrSessionUnavailable = errors.New("session store unavailable")

func NewService(repo Repository, r *redis.Client, s Sender, secret string, access, refresh time.Duration, metrics ...Metrics) *Service {
	service := &Service{repo: repo, redis: r, sender: s, secret: []byte(secret), accessTTL: access, refreshTTL: refresh}
	if len(metrics) > 0 {
		service.metrics = metrics[0]
	}
	return service
}
func normalizePhone(v string) (string, error) {
	v = strings.ReplaceAll(strings.TrimSpace(v), " ", "")
	if strings.HasPrefix(v, "+86") {
		v = v[3:]
	}
	if len(v) != 11 || v[0] != '1' {
		return "", errors.New("invalid phone")
	}
	for _, r := range v {
		if r < '0' || r > '9' {
			return "", errors.New("invalid phone")
		}
	}
	return "+86" + v, nil
}
func validDevice(v string) bool { n := len(strings.TrimSpace(v)); return n >= 8 && n <= 200 }
func (s *Service) codeMAC(code string) string {
	m := hmac.New(sha256.New, s.secret)
	_, _ = m.Write([]byte(code))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

var rateScript = redis.NewScript(`
local limits={tonumber(ARGV[1]),tonumber(ARGV[2]),tonumber(ARGV[3]),tonumber(ARGV[4])}
local ttls={tonumber(ARGV[5]),tonumber(ARGV[6]),tonumber(ARGV[7]),tonumber(ARGV[8])}
for i=1,4 do local v=tonumber(redis.call('GET',KEYS[i]) or '0');if v>=limits[i] then return 0 end end
for i=1,4 do local v=redis.call('INCR',KEYS[i]);if v==1 then redis.call('EXPIRE',KEYS[i],ttls[i]) end end
return 1`)

var verifyScript = redis.NewScript(`
local raw=redis.call('GET',KEYS[1]);if not raw then return '' end
local c=cjson.decode(raw);if c.deviceId~=ARGV[1] then return '' end
if c.codeMac~=ARGV[2] then c.attempts=(c.attempts or 0)+1;if c.attempts>=5 then redis.call('DEL',KEYS[1]) else redis.call('SET',KEYS[1],cjson.encode(c),'KEEPTTL') end;return '' end
redis.call('DEL',KEYS[1]);return c.phone`)

func (s *Service) Send(ctx context.Context, phone, deviceID, clientIP string) (string, error) {
	phone, err := normalizePhone(phone)
	if err != nil || !validDevice(deviceID) || clientIP == "" {
		s.observe("sms_send", "invalid")
		return "", ErrInvalidRequest
	}
	day := time.Now().UTC().Format("20060102")
	keys := []string{"otp:cooldown:" + phone, "otp:phone-day:" + phone + ":" + day, "otp:device-day:" + deviceID + ":" + day, "otp:ip-hour:" + clientIP + ":" + time.Now().UTC().Format("2006010215")}
	ok, err := rateScript.Run(ctx, s.redis, keys, 1, 5, 10, 20, 60, 86400, 86400, 3600).Int()
	if err != nil {
		s.observe("sms_send", "redis_error")
		return "", err
	}
	if ok != 1 {
		s.observe("sms_send", "rate_limited")
		return "", ErrRateLimited
	}
	raw := make([]byte, 4)
	if _, err = rand.Read(raw); err != nil {
		s.observe("sms_send", "internal_error")
		return "", err
	}
	code := fmt.Sprintf("%06d", (int(raw[0])<<16|int(raw[1])<<8|int(raw[2]))%1000000)
	id := uuid.NewString()
	payload, _ := json.Marshal(challenge{Phone: phone, DeviceID: deviceID, CodeMAC: s.codeMAC(code)})
	if err = s.redis.Set(ctx, "otp:challenge:"+id, payload, 5*time.Minute).Err(); err != nil {
		s.observe("sms_send", "redis_error")
		return "", err
	}
	if err = s.sender.SendCode(ctx, phone, code); err != nil {
		_ = s.redis.Del(ctx, "otp:challenge:"+id).Err()
		s.observe("sms_send", "provider_error")
		return "", err
	}
	s.observe("sms_send", "success")
	return id, nil
}

func (s *Service) Verify(ctx context.Context, challengeID, code, deviceID string) (Tokens, error) {
	if _, err := uuid.Parse(challengeID); err != nil || !validDevice(deviceID) || ValidateCode(code) != nil {
		s.observe("sms_verify", "rejected")
		return Tokens{}, ErrInvalidChallenge
	}
	phone, err := verifyScript.Run(ctx, s.redis, []string{"otp:challenge:" + challengeID}, deviceID, s.codeMAC(code)).Text()
	if err != nil {
		s.observe("sms_verify", "redis_error")
		return Tokens{}, fmt.Errorf("%w: %v", ErrVerificationUnavailable, err)
	}
	if phone == "" {
		s.observe("sms_verify", "rejected")
		return Tokens{}, ErrInvalidChallenge
	}
	userID, err := s.repo.FindOrCreateUser(ctx, phone)
	if err != nil {
		s.observe("sms_verify", "repository_error")
		return Tokens{}, err
	}
	sessionID := uuid.NewString()
	tokens, hash, err := s.issue(userID, sessionID)
	if err != nil {
		s.observe("sms_verify", "internal_error")
		return Tokens{}, err
	}
	if err = s.repo.CreateSession(ctx, sessionID, userID, hash, deviceID, tokens.RefreshExpiresAt); err != nil {
		s.observe("sms_verify", "repository_error")
		return Tokens{}, err
	}
	s.observe("sms_verify", "success")
	return tokens, nil
}

func (s *Service) observe(operation, outcome string) {
	if s.metrics != nil {
		s.metrics.ObserveAuth(operation, outcome)
	}
}

func (s *Service) issue(userID, sessionID string) (Tokens, []byte, error) {
	now := time.Now().UTC()
	accessExp := now.Add(s.accessTTL)
	claims := jwt.MapClaims{"sub": userID, "sid": sessionID, "iat": now.Unix(), "exp": accessExp.Unix(), "typ": "access"}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return Tokens{}, nil, err
	}
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return Tokens{}, nil, err
	}
	refresh := sessionID + "." + base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(refresh))
	return Tokens{AccessToken: access, RefreshToken: refresh, AccessExpiresAt: accessExp, RefreshExpiresAt: now.Add(s.refreshTTL)}, hash[:], nil
}
func (s *Service) ParseAccess(ctx context.Context, raw string, requireSession bool) (string, string, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["typ"] != "access" {
		return "", "", errors.New("invalid token")
	}
	sub, e := claims.GetSubject()
	sid, _ := claims["sid"].(string)
	if e != nil || sub == "" || sid == "" {
		return "", "", errors.New("invalid token")
	}
	if requireSession {
		active, repoErr := s.repo.IsActiveSession(ctx, sid, sub)
		if repoErr != nil {
			return "", "", fmt.Errorf("%w: %v", ErrSessionUnavailable, repoErr)
		}
		if !active {
			return "", "", errors.New("invalid token")
		}
	}
	return sub, sid, nil
}
func (s *Service) Refresh(ctx context.Context, raw string) (Tokens, error) {
	parts := strings.SplitN(raw, ".", 2)
	if len(parts) != 2 {
		return Tokens{}, errors.New("invalid refresh token")
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		return Tokens{}, errors.New("invalid refresh token")
	}
	hash := sha256.Sum256([]byte(raw))
	return s.repo.RotateSession(ctx, parts[0], hash[:], func(userID string) (Tokens, []byte, error) { return s.issue(userID, parts[0]) })
}
func (s *Service) Logout(ctx context.Context, raw string) error {
	parts := strings.SplitN(raw, ".", 2)
	if len(parts) != 2 {
		return errors.New("invalid refresh token")
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		return errors.New("invalid refresh token")
	}
	revoked, err := s.repo.RevokeSession(ctx, parts[0])
	if err != nil {
		return err
	}
	if !revoked {
		return errors.New("invalid refresh token")
	}
	return nil
}
func ValidateCode(v string) error {
	if len(v) != 6 {
		return errors.New("invalid code")
	}
	for _, character := range v {
		if character < '0' || character > '9' {
			return errors.New("invalid code")
		}
	}
	return nil
}
