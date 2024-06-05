package repo

import "context"

func (r *Repo) UpdateLinkUsageInfo(ctx context.Context, id int64) error {
	err := r.q.UpdateLinkUsageInfo(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
