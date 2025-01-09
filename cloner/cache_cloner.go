package cloner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drone-plugins/drone-github-actions/cache"
	"golang.org/x/exp/slog"
)

func NewCache(cloner Cloner) *cacheCloner {
	return &cacheCloner{cloner: cloner}
}

type cacheCloner struct {
	cloner Cloner
}

// Clone method clones the repository & caches it if not present in cache already.
func (c *cacheCloner) Clone(ctx context.Context, repo, ref, sha string) (string, error) {
	key := cache.GetKeyName(fmt.Sprintf("%s%s%s", repo, ref, sha))
	codedir := filepath.Join(key, "data")

	cloneFn := func() error {
		// Remove stale data
		if err := os.RemoveAll(codedir); err != nil {
			slog.Error("cannot remove code directory", codedir, err)
		}

		if err := os.MkdirAll(codedir, 0700); err != nil {
			slog.Error("failed to create code directory", codedir, err)
			return err
		}
		return c.cloner.Clone(ctx,
			Params{Repo: repo, Ref: ref, Sha: sha, Dir: codedir})
	}

	if err := cache.Add(key, cloneFn); err != nil {
		return "", err
	}
	return codedir, nil
}
