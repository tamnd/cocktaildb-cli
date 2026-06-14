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
lookup by ID, and category listing.`,
			Site: Host,
			Repo: "https://github.com/tamnd/cocktaildb-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// search: find drinks by name
	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "read",
		List:    true,
		Summary: "Search cocktails by name",
		Args:    []kit.Arg{{Name: "query", Help: "cocktail name to search"}},
	}, searchOp)

	// lookup: fetch drink by ID
	kit.Handle(app, kit.OpMeta{
		Name:    "lookup",
		Group:   "read",
		Single:  true,
		Summary: "Fetch a cocktail by ID",
		Args:    []kit.Arg{{Name: "id", Help: "drink ID"}},
	}, lookupOp)

	// random: one random drink
	kit.Handle(app, kit.OpMeta{
		Name:    "random",
		Group:   "read",
		Single:  true,
		Summary: "Fetch a random cocktail",
	}, randomOp)

	// categories: list categories, alcoholic types, glasses, or ingredients
	kit.Handle(app, kit.OpMeta{
		Name:    "categories",
		Group:   "read",
		List:    true,
		Summary: "List categories, alcoholic types, glass types, or ingredients",
	}, categoriesOp)
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
	Query  string        `kit:"arg"          help:"cocktail name to search"`
	Limit  int           `kit:"flag,inherit" help:"max results"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type lookupInput struct {
	ID     string  `kit:"arg" help:"drink ID"`
	Client *Client `kit:"inject"`
}

type randomInput struct {
	Client *Client `kit:"inject"`
}

type categoriesInput struct {
	Type   string  `kit:"flag" help:"list type: categories|alcoholic|glass|ingredients" default:"categories"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func searchOp(ctx context.Context, in searchInput, emit func(Drink) error) error {
	items, err := in.Client.Search(ctx, in.Query, in.Limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func lookupOp(ctx context.Context, in lookupInput, emit func(Drink) error) error {
	drink, err := in.Client.Lookup(ctx, in.ID)
	if err != nil {
		return err
	}
	return emit(drink)
}

func randomOp(ctx context.Context, in randomInput, emit func(Drink) error) error {
	drink, err := in.Client.Random(ctx)
	if err != nil {
		return err
	}
	return emit(drink)
}

func categoriesOp(ctx context.Context, in categoriesInput, emit func(Category) error) error {
	listType := in.Type
	if listType == "" {
		listType = "categories"
	}
	cats, err := in.Client.ListCategories(ctx, listType)
	if err != nil {
		return err
	}
	for _, cat := range cats {
		if err := emit(cat); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
// Numeric inputs are treated as IDs; others as search queries.
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty cocktaildb reference")
	}
	isNumeric := true
	for _, ch := range input {
		if ch < '0' || ch > '9' {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		return "id", input, nil
	}
	return "query", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "id":
		return "https://www.thecocktaildb.com/drink/" + id, nil
	case "query":
		return "https://www.thecocktaildb.com/drink/" + id + "-Detail.php", nil
	default:
		return "", errs.Usage("cocktaildb has no resource type %q", uriType)
	}
}
