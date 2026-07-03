package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"wa-assistant/backend/config"
)

const (
	StatusActive          = "active"
	StatusRevoked         = "revoked"
	StatusExpired         = "expired"
	StatusMachineMismatch = "machine_mismatch"
	StatusMachineLimit    = "machine_limit_reached"
	StatusNotFound        = "not_found"
	StatusNetworkError    = "network_error"
	StatusRateLimited     = "rate_limited"
	StatusServerError     = "server_error"
	StatusUnauthorized    = "unauthorized"
	StatusInvalidRequest  = "invalid_request"
	StatusGraceExpired    = "offline_grace_expired"
)

type verifyRequest struct {
	LicenseKey        string `json:"license_key"`
	MachineHash       string `json:"machine_hash"`
	LegacyMachineHash string `json:"legacy_machine_hash,omitempty"`
	Heartbeat         bool   `json:"heartbeat,omitempty"`
}

type verifyResponse struct {
	Valid          bool   `json:"valid"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	PackageType    string `json:"package_type"`
	ResetRemaining int    `json:"reset_remaining"`
	MachinesUsed   int    `json:"machines_used"`
	MachineMax     int    `json:"machine_max"`
}

type successWrapper struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    *verifyResponse `json:"data"`
}

type verificationResult struct {
	Valid          bool
	Status         string
	Message        string
	PackageType    string
	ResetRemaining int
	MachinesUsed   int
	MachineMax     int
}

var (
	stateMu sync.RWMutex

	VerifyResult      bool
	VerifyMessage     string
	VerifyStatus      string
	VerifyPackageType string
	lastHeartbeatOK   = true
	lastHeartbeatAt   time.Time
	heartbeatFailCnt  int

	licenseHTTPClient = &http.Client{Timeout: 15 * time.Second}
)

// DevMode can only be enabled at build time with:
// -ldflags "-X wa-assistant/backend/license.DevMode=true".
var DevMode = "false"

// Verify validates and binds the license before the application starts.
func Verify() bool {
	if DevMode == "true" {
		setVerificationState(true, "dev", "dev mode", "")
		log.Println("[license] DEV MODE — skip verifikasi lisensi")
		return true
	}

	key := strings.TrimSpace(config.Env("LICENSE_KEY", ""))
	if key == "" {
		setVerificationState(false, "no_key", "LICENSE_KEY kosong. Beli lisensi di ngertikode.id", "")
		log.Printf("[license] GAGAL: %s", VerifyMessage)
		return false
	}

	machine, legacyMachine, err := machineFingerprints()
	if err != nil {
		setVerificationState(false, "machine_id_error", fmt.Sprintf("gagal menyiapkan identitas instalasi: %v", err), "")
		log.Printf("[license] GAGAL: %s", VerifyMessage)
		return false
	}

	log.Printf("[license] Verifying key=%s machine=%s...", maskKey(key), machine[:8])
	result := callVerifyAPI(key, machine, legacyMachine, false)
	setVerificationState(result.Valid, result.Status, result.Message, result.PackageType)
	if !result.Valid {
		log.Printf("[license] VERIFY GAGAL status=%s: %s", result.Status, result.Message)
		return false
	}

	stateMu.Lock()
	lastHeartbeatOK = true
	lastHeartbeatAt = time.Now()
	heartbeatFailCnt = 0
	stateMu.Unlock()
	log.Printf("[license] VERIFY OK — status=%s package=%s", result.Status, result.PackageType)
	return true
}

// Heartbeat revalidates the license. Terminal server decisions stop the app
// immediately; connectivity/server failures are tolerated only for the
// configured offline grace period.
func Heartbeat() bool {
	if DevMode == "true" {
		return true
	}

	key := strings.TrimSpace(config.Env("LICENSE_KEY", ""))
	if key == "" {
		setVerificationState(false, "no_key", "LICENSE_KEY kosong", "")
		return false
	}

	machine, legacyMachine, err := machineFingerprints()
	if err != nil {
		setVerificationState(false, "machine_id_error", fmt.Sprintf("gagal membaca identitas instalasi: %v", err), "")
		return false
	}

	result := callVerifyAPI(key, machine, legacyMachine, true)
	now := time.Now()

	stateMu.Lock()
	defer stateMu.Unlock()
	if result.Valid {
		VerifyResult = true
		VerifyMessage = result.Message
		VerifyStatus = result.Status
		VerifyPackageType = result.PackageType
		lastHeartbeatOK = true
		lastHeartbeatAt = now
		heartbeatFailCnt = 0
		return true
	}

	heartbeatFailCnt++
	lastHeartbeatOK = false
	log.Printf("[license] Heartbeat gagal (x%d) status=%s: %s", heartbeatFailCnt, result.Status, result.Message)

	if isTerminalStatus(result.Status) {
		VerifyResult = false
		VerifyMessage = result.Message
		VerifyStatus = result.Status
		return false
	}

	grace := offlineGraceDuration()
	if lastHeartbeatAt.IsZero() || !now.Before(lastHeartbeatAt.Add(grace)) {
		VerifyResult = false
		VerifyStatus = StatusGraceExpired
		VerifyMessage = fmt.Sprintf("server lisensi tidak dapat diverifikasi selama %s: %s", grace, result.Message)
		return false
	}

	return true
}

func callVerifyAPI(key, machine, legacyMachine string, heartbeat bool) verificationResult {
	baseURL := strings.TrimRight(config.Env("LICENSE_API_URL", "https://api.ngertikode.id"), "/")
	body := verifyRequest{
		LicenseKey:        key,
		MachineHash:       machine,
		LegacyMachineHash: legacyMachine,
		Heartbeat:         heartbeat,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return verificationResult{Status: StatusInvalidRequest, Message: fmt.Sprintf("gagal menyusun request lisensi: %v", err)}
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/license/verify", bytes.NewReader(bodyJSON))
	if err != nil {
		return verificationResult{Status: StatusInvalidRequest, Message: fmt.Sprintf("gagal membuat request: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := strings.TrimSpace(config.Env("LICENSE_API_SECRET", "")); secret != "" {
		req.Header.Set("X-License-Secret", secret)
	}

	resp, err := licenseHTTPClient.Do(req)
	if err != nil {
		return verificationResult{Status: StatusNetworkError, Message: fmt.Sprintf("server tidak terjangkau: %v", err)}
	}
	defer resp.Body.Close()

	var wrapper successWrapper
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return verificationResult{Status: statusFromHTTP(resp.StatusCode), Message: fmt.Sprintf("respons server tidak valid (HTTP %d)", resp.StatusCode)}
	}
	if wrapper.Data == nil {
		message := strings.TrimSpace(wrapper.Message)
		if message == "" {
			message = "server menolak verifikasi"
		}
		return verificationResult{Status: statusFromHTTP(resp.StatusCode), Message: message}
	}

	vr := wrapper.Data
	status := strings.TrimSpace(vr.Status)
	if status == "" {
		status = inferLegacyStatus(vr.Valid, vr.Message)
	}
	return verificationResult{
		Valid:          vr.Valid,
		Status:         status,
		Message:        vr.Message,
		PackageType:    vr.PackageType,
		ResetRemaining: vr.ResetRemaining,
		MachinesUsed:   vr.MachinesUsed,
		MachineMax:     vr.MachineMax,
	}
}

func statusFromHTTP(statusCode int) string {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return StatusUnauthorized
	case http.StatusUnprocessableEntity, http.StatusBadRequest:
		return StatusInvalidRequest
	case http.StatusTooManyRequests:
		return StatusRateLimited
	default:
		return StatusServerError
	}
}

func inferLegacyStatus(valid bool, message string) string {
	if valid {
		return StatusActive
	}
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "revoked"), strings.Contains(lower, "suspend"):
		return StatusRevoked
	case strings.Contains(lower, "expired"):
		return StatusExpired
	case strings.Contains(lower, "mismatch"):
		return StatusMachineMismatch
	case strings.Contains(lower, "machine limit"), strings.Contains(lower, "limit reached"):
		return StatusMachineLimit
	case strings.Contains(lower, "not found"):
		return StatusNotFound
	default:
		return StatusServerError
	}
}

func isTerminalStatus(status string) bool {
	switch status {
	case StatusRevoked, StatusExpired, StatusMachineMismatch, StatusMachineLimit, StatusNotFound, StatusUnauthorized, StatusInvalidRequest:
		return true
	default:
		return false
	}
}

func offlineGraceDuration() time.Duration {
	hours := config.EnvInt("LICENSE_OFFLINE_GRACE_HOURS", 24)
	if hours < 1 {
		hours = 1
	}
	return time.Duration(hours) * time.Hour
}

func setVerificationState(valid bool, status, message, packageType string) {
	stateMu.Lock()
	VerifyResult = valid
	VerifyStatus = status
	VerifyMessage = message
	VerifyPackageType = packageType
	stateMu.Unlock()
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "***"
}
