package app

import "github.com/dector/lampa/server/core/logstore"

func webUIStatsSnapshot(logs logstore.Store) (int, int64) {
	if logs == nil {
		return 0, 0
	}
	return logs.Count(), logs.SizeBytes()
}
