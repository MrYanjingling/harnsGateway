package options

import (
	"fmt"
	"strings"

	"harns-platform/pkg/audit"
)

func Validate(o *Options) []error {
	var errs []error
	if err := o.BaseOptions.ValidateAndApply(); err != nil {
		errs = append(errs, err)
	}

	if strings.TrimSpace(o.Port) == "" {
		errs = append(errs, fmt.Errorf("port must not be empty"))
	}
	if o.Wait <= 0 {
		errs = append(errs, fmt.Errorf("graceful-timeout must be greater than 0"))
	}
	if o.MaxBodyBytes <= 0 {
		errs = append(errs, fmt.Errorf("max-body-bytes must be greater than 0"))
	}
	if len(o.CORS.AllowedOrigins) == 0 {
		errs = append(errs, fmt.Errorf("cors.allowed-origins must not be empty"))
	}
	if len(o.CORS.AllowedMethods) == 0 {
		errs = append(errs, fmt.Errorf("cors.allowed-methods must not be empty"))
	}
	if o.CORS.MaxAgeSeconds < 0 {
		errs = append(errs, fmt.Errorf("cors.max-age-seconds must not be negative"))
	}

	if len(o.Redis.Addrs) == 0 {
		errs = append(errs, fmt.Errorf("redis addrs must not be empty"))
	}
	if o.Redis.ReadTimeoutSeconds < 0 || o.Redis.WriteTimeoutSeconds < 0 {
		errs = append(errs, fmt.Errorf("redis timeouts must not be negative"))
	}

	if strings.TrimSpace(o.MySQL.Host) == "" || o.MySQL.Port <= 0 || strings.TrimSpace(o.MySQL.Database) == "" {
		errs = append(errs, fmt.Errorf("mysql host/port/database must be set"))
	}
	if o.MySQL.MaxOpenConns <= 0 {
		errs = append(errs, fmt.Errorf("mysql max-open-conns must be greater than 0"))
	}
	if o.MySQL.MaxIdleConns < 0 {
		errs = append(errs, fmt.Errorf("mysql max-idle-conns must not be negative"))
	}
	if o.MySQL.ConnMaxLifetimeMinutes <= 0 {
		errs = append(errs, fmt.Errorf("mysql conn-max-lifetime-minutes must be greater than 0"))
	}

	if strings.TrimSpace(o.MinIO.Endpoint) == "" || strings.TrimSpace(o.MinIO.Bucket) == "" {
		errs = append(errs, fmt.Errorf("minio endpoint/bucket must be set"))
	}

	if strings.TrimSpace(o.Security.TokenHeader) == "" {
		errs = append(errs, fmt.Errorf("security.token-header must not be empty"))
	}
	if len(o.Security.TokenSecret) < 32 {
		errs = append(errs, fmt.Errorf("security.token-secret must be at least 32 bytes"))
	}
	if o.Security.ExpireMinutes <= 0 {
		errs = append(errs, fmt.Errorf("security.expire-minutes must be greater than 0"))
	}
	if o.Security.RefreshMinutes < 0 {
		errs = append(errs, fmt.Errorf("security.refresh-minutes must not be negative"))
	}
	if o.Security.MaxLoginFailed <= 0 {
		errs = append(errs, fmt.Errorf("security.max-login-failed must be greater than 0"))
	}
	if o.Security.LockMinutes <= 0 {
		errs = append(errs, fmt.Errorf("security.lock-minutes must be greater than 0"))
	}
	if o.Security.CaptchaTTLSeconds <= 0 {
		errs = append(errs, fmt.Errorf("security.captcha-ttl-seconds must be greater than 0"))
	}

	if o.Telemetry.Enabled && strings.TrimSpace(o.Telemetry.Endpoint) == "" {
		errs = append(errs, fmt.Errorf("telemetry.endpoint must not be empty when enabled"))
	}

	if o.Audit.Enabled {
		if o.Audit.MaxBodyBytes <= 0 {
			errs = append(errs, fmt.Errorf("audit.max-body-bytes must be greater than 0"))
		}
		if o.Audit.QueueSize <= 0 {
			errs = append(errs, fmt.Errorf("audit.queue-size must be greater than 0"))
		}
		if o.Audit.RetentionDays <= 0 {
			errs = append(errs, fmt.Errorf("audit.retention-days must be greater than 0"))
		}
		if o.Audit.CleanupBatchSize <= 0 {
			errs = append(errs, fmt.Errorf("audit.cleanup-batch-size must be greater than 0"))
		}
		if o.Audit.BatchSize <= 0 || o.Audit.BatchSize > audit.MaxBatchSize {
			errs = append(errs, fmt.Errorf("audit.batch-size must be between 1 and %d", audit.MaxBatchSize))
		}
		if o.Audit.FlushIntervalMs <= 0 {
			errs = append(errs, fmt.Errorf("audit.flush-interval-ms must be greater than 0"))
		}
	}

	return errs
}
