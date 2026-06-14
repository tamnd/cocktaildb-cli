package cocktaildb

import (
	"context"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes cocktaildb as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/cocktaildb-cli/cocktaildb"
//
// The same Domain also builds the standalone cocktaildb binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the cocktaildb driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "cocktaildb",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "cocktaildb",
			Short:  "TheCocktailDB cocktail search and browser",
			Long: `cocktaildb fetches cocktail recipes and categories from TheCocktailDB
public API. No API key required. Supports name search, random picks,
and category listing.`,
			Site: Host,
			Repo: "https://github.com/tamnd/cocktaildb-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// search: find cocktails by name
	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "read",
		List:    true,
		Summary: "Search cocktails by name",
		Args:    []kit.Arg{{Name: "name", Help: "cocktail name to search for"}},
	}, searchOp)

	// random: one random cocktail
	kit.Handle(app, kit.OpMeta{
		Name:    "random",
		Group:   "read",
		Single:  true,
		Summary: "Fetch a random cocktail",
	}, randomOp)

	// get: fetch cocktail by ID
	kit.Handle(app, kit.OpMeta{
		Name:    "get",
		Group:   "read",
		Single:  true,
		Summary: "Fetch a cocktail by ID",
		Args:    []kit.Arg{{Name: "id", Help: "cocktail ID"}},
	}, getOp)

	// filter: filter cocktails by alcoholic, category, or glass
	kit.Handle(app, kit.OpMeta{
		Name:    "filter",
		Group:   "read",
		List:    true,
		Summary: "Filter cocktails by alcoholic status, category, or glass type",
	}, filterOp)

	// categories: list all cocktail categories
	kit.Handle(app, kit.OpMeta{
		Name:    "categories",
		Group:   "read",
		List:    true,
		Summary: "List all cocktail categories",
	}, categoriesOp)

	// glasses: list all glass types
	kit.Handle(app, kit.OpMeta{
		Name:    "glasses",
		Group:   "read",
		List:    true,
		Summary: "List all glass types",
	}, glassesOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type searchInput struct {
	Name   string        `kit:"arg"          help:"cocktail name to search for"`
	Limit  int           `kit:"flag,inherit" help:"max results"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type randomInput struct {
	Client *Client `kit:"inject"`
}

type getInput struct {
	ID     string  `kit:"arg" help:"cocktail ID"`
	Client *Client `kit:"inject"`
}

type filterInput struct {
	Alcoholic string  `kit:"flag" help:"filter by alcoholic status (Alcoholic, Non_Alcoholic, Optional_Alcohol)"`
	Category  string  `kit:"flag" help:"filter by category"`
	Glass     string  `kit:"flag" help:"filter by glass type"`
	Limit     int     `kit:"flag,inherit" help:"max results"`
	Client    *Client `kit:"inject"`
}

type categoriesInput struct {
	Client *Client `kit:"inject"`
}

type glassesInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func searchOp(ctx context.Context, in searchInput, emit func(Cocktail) error) error {
	items, err := in.Client.Search(ctx, in.Name, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func randomOp(ctx context.Context, in randomInput, emit func(Cocktail) error) error {
	cocktail, err := in.Client.Random(ctx)
	if err != nil {
		return mapErr(err)
	}
	return emit(cocktail)
}

func getOp(ctx context.Context, in getInput, emit func(Cocktail) error) error {
	cocktail, err := in.Client.Get(ctx, in.ID)
	if err != nil {
		return mapErr(err)
	}
	return emit(cocktail)
}

func filterOp(ctx context.Context, in filterInput, emit func(FilterResult) error) error {
	opts := FilterOptions{
		Alcoholic: in.Alcoholic,
		Category:  in.Category,
		Glass:     in.Glass,
		Limit:     in.Limit,
	}
	results, err := in.Client.Filter(ctx, opts)
	if err != nil {
		return mapErr(err)
	}
	for _, r := range results {
		if err := emit(r); err != nil {
			return err
		}
	}
	return nil
}

func categoriesOp(ctx context.Context, in categoriesInput, emit func(Category) error) error {
	cats, err := in.Client.Categories(ctx)
	if err != nil {
		return mapErr(err)
	}
	for _, cat := range cats {
		if err := emit(cat); err != nil {
			return err
		}
	}
	return nil
}

func glassesOp(ctx context.Context, in glassesInput, emit func(GlassType) error) error {
	glasses, err := in.Client.Glasses(ctx)
	if err != nil {
		return mapErr(err)
	}
	for _, g := range glasses {
		if err := emit(g); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty cocktaildb reference")
	}
	return "cocktail", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "cocktail":
		return "https://www.thecocktaildb.com/drink/" + id, nil
	default:
		return "", errs.Usage("cocktaildb has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
