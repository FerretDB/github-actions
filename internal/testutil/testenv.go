package testutil

import "testing"

func GetEnvFunc(t *testing.T, env map[string]string) func(string) string {
	return func(key string) string {
		if val, ok := env[key]; ok {
			return val
		}

		t.Fatalf("unexpected key %q", key)
		panic("not reached")
	}
}
