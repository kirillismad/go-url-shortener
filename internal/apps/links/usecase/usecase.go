package usecase

import "context"

type ICreateLinkHandler interface {
	CreateLink(ctx context.Context, data CreateLinkData) (CreateLinkResult, error)
}

type IGetLinkByShortIDHandler interface {
	GetLinkByShortID(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error)
}

type ILinksFacade interface {
	ICreateLinkHandler
	IGetLinkByShortIDHandler
}

type LinksFacade struct {
	createLinkHandler       ICreateLinkHandler
	getLinkByShortIDHandler IGetLinkByShortIDHandler
}

type CreateLinkData struct {
}

type CreateLinkResult struct {
}

func (f *LinksFacade) CreateLink(ctx context.Context, data CreateLinkData) (CreateLinkResult, error) {
	return f.createLinkHandler.CreateLink(ctx, data)
}

type GetLinkByShortIDData struct {
}

type GetLinkByShortIDResult struct {
}

func (f *LinksFacade) GetLinkByShortID(ctx context.Context, data GetLinkByShortIDData) (GetLinkByShortIDResult, error) {
	return f.getLinkByShortIDHandler.GetLinkByShortID(ctx, data)
}
