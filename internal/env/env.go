package env

import "github.com/codingconcepts/env"

type Env struct {
	MaxConcurrentReconciles int `env:"MAX_CONCURRENT_RECONCILES" default:"5"`

	IACJobsNamespace string `env:"IAC_JOBS_NAMESPACE" required:"true"`
	IACJobImage      string `env:"IAC_JOB_IMAGE" required:"true"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
