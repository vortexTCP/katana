package runner

import (
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/katana/pkg/engine"
	"github.com/projectdiscovery/katana/pkg/engine/hybrid"
	"github.com/projectdiscovery/katana/pkg/engine/standard"
	"github.com/projectdiscovery/katana/pkg/types"
	fileutil "github.com/projectdiscovery/utils/file"
	"go.uber.org/multierr"
)

// Runner creates the required resources for crawling
// and executes the crawl process.
type Runner struct {
	crawlerOptions *types.CrawlerOptions
	stdin          bool
	crawler        engine.Engine
	options        *types.Options
}

// New returns a new crawl runner structure
func New(options *types.Options) (*Runner, error) {
	configureOutput(options)
	showBanner()

	if options.Version {
		gologger.Info().Msgf("Current version: %s", version)
		return nil, nil
	}

	if err := initExampleFormFillConfig(); err != nil {
		return nil, errors.Wrap(err, "could not init default config")
	}
	if err := validateOptions(options); err != nil {
		return nil, errors.Wrap(err, "could not validate options")
	}
	if options.FormConfig != "" {
		if err := readCustomFormConfig(options); err != nil {
			return nil, err
		}
	}
	crawlerOptions, err := types.NewCrawlerOptions(options)
	if err != nil {
		return nil, errors.Wrap(err, "could not create crawler options")
	}

	var (
		crawler engine.Engine
	)

	switch {
	case options.Headless:
		crawler, err = hybrid.New(crawlerOptions)
	default:
		crawler, err = standard.New(crawlerOptions)
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not create standard crawler")
	}
	runner := &Runner{options: options, stdin: fileutil.HasStdin(), crawlerOptions: crawlerOptions, crawler: crawler}

	return runner, nil
}

// Close closes the runner releasing resources
func (r *Runner) Close() error {
	return multierr.Combine(
		r.crawler.Close(),
		r.crawlerOptions.Close(),
	)
}
