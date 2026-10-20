package options

import "fmt"

func Validate(o *Options) []error {
	var errs []error
	if err := o.BaseOptions.ValidateAndApply(); err != nil {
		errs = append(errs, err)
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

	return errs
}
