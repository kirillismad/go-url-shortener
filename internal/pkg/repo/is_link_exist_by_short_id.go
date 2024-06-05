package repo

import "context"

func (r *Repo) IsLinkExistByShortID(ctx context.Context, shortID string) (bool, error) {
	exist, err := r.q.IsLinkExistByShortID(ctx, shortID)
	if err != nil {
		return false, err
	}
	return exist, nil
}
