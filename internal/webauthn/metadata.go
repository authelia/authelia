package webauthn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/go-webauthn/webauthn/metadata/providers/cached"
	"github.com/go-webauthn/webauthn/metadata/providers/memory"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

// NewMetaDataProvider generates a new metadata.Provider given a *schema.Configuration and storage.CachedDataProvider.
func NewMetaDataProvider(config *schema.Configuration, store storage.CachedDataProvider) (provider MetaDataProvider, err error) {
	if config.WebAuthn.Metadata.Enabled {
		p := &StoreCachedMetadataProvider{
			new:         newMetadataProviderMemory(config),
			clock:       &metadata.RealClock{},
			store:       store,
			log:         logging.Logger().WithFields(map[string]any{logging.FieldProvider: "webauthn-metadata"}),
			cachePolicy: config.WebAuthn.Metadata.CachePolicy,
			handler:     &productionMDS3Provider{},
		}

		if p.decoder, err = metadata.NewDecoder(metadata.WithIgnoreEntryParsingErrors()); err != nil {
			return nil, err
		}

		provider = p
	}

	return provider, nil
}

func newMetadataProviderMemory(config *schema.Configuration) cached.NewFunc {
	return func(mds *metadata.Metadata) (provider metadata.Provider, err error) {
		return memory.New(
			memory.WithMetadata(mds.ToMap()),
			memory.WithValidateEntry(config.WebAuthn.Metadata.ValidateEntry),
			memory.WithValidateEntryPermitZeroAAGUID(config.WebAuthn.Metadata.ValidateEntryPermitZeroAAGUID),
			memory.WithValidateTrustAnchor(config.WebAuthn.Metadata.ValidateTrustAnchor),
			memory.WithValidateStatus(config.WebAuthn.Metadata.ValidateStatus),
			memory.WithStatusUndesired(config.WebAuthn.Metadata.ValidateStatusProhibited),
			memory.WithStatusDesired(config.WebAuthn.Metadata.ValidateStatusPermitted),
		)
	}
}

type MetaDataProvider interface {
	metadata.Provider

	StartupCheck() (err error)
	Load(ctx context.Context) (mds *metadata.Metadata, data []byte, err error)
	LoadForce(ctx context.Context) (mds *metadata.Metadata, data []byte, err error)
	LoadFile(ctx context.Context, path string) (mds *metadata.Metadata, data []byte, err error)
	LoadCache(ctx context.Context) (mds *metadata.Metadata, data []byte, err error)
	SaveCache(ctx context.Context, data []byte) (err error)
	Outdated() (outdated bool)
}

type StoreCachedMetadataProvider struct {
	metadata.Provider

	new func(mds *metadata.Metadata) (provider metadata.Provider, err error)

	mu      sync.Mutex
	store   storage.CachedDataProvider
	decoder *metadata.Decoder
	clock   metadata.Clock
	handler MDS3Provider

	log *logrus.Entry

	cachePolicy string

	update time.Time
	number int
}

func (p *StoreCachedMetadataProvider) StartupCheck() (err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.init()
}

func (p *StoreCachedMetadataProvider) GetEntry(ctx context.Context, aaguid uuid.UUID) (entry *metadata.Entry, err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	if p.outdated() {
		var (
			mds  *metadata.Metadata
			data []byte
		)

		if mds, data, err = p.getCache(ctx); err != nil {
			if mds, data, err = p.get(ctx, -1); err != nil {
				return nil, err
			}
		}

		if err = p.configure(mds); err != nil {
			return nil, err
		}

		if err = p.saveCache(ctx, data); err != nil {
			return nil, err
		}
	}

	return p.Provider.GetEntry(ctx, aaguid)
}

func (p *StoreCachedMetadataProvider) init() (err error) {
	if p.store == nil {
		return fmt.Errorf("error initializing provider: storage is not configured")
	}

	var data []byte

	ctx := context.Background()

	_, _, _ = p.loadCache(ctx)

	if _, data, err = p.loadCurrent(ctx, p.number); err != nil {
		if p.cachePolicy == CachePolicyStrict {
			return fmt.Errorf("error initializing provider: %w", err)
		}

		p.log.WithError(err).Debug("Error occurred fetching current metadata but the cache policy is relaxed")
	}

	if p.number <= 0 {
		return fmt.Errorf("error initializing provider: no metadata was loaded")
	}

	if p.update.Before(p.clock.Now()) {
		return fmt.Errorf("error initializing provider: outdated metadata was loaded")
	}

	if data == nil {
		return nil
	}

	return p.saveCache(ctx, data)
}

func (p *StoreCachedMetadataProvider) Load(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.load(ctx)
}

func (p *StoreCachedMetadataProvider) LoadForce(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.loadForced(ctx)
}

func (p *StoreCachedMetadataProvider) load(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	return p.loadCurrent(ctx, p.number)
}

func (p *StoreCachedMetadataProvider) loadForced(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	return p.loadCurrent(ctx, -1)
}

func (p *StoreCachedMetadataProvider) loadCurrent(ctx context.Context, current int) (mds *metadata.Metadata, data []byte, err error) {
	if mds, data, err = p.get(ctx, current); err != nil {
		return nil, nil, err
	}

	if err = p.configure(mds); err != nil {
		return nil, nil, err
	}

	return mds, data, nil
}

func (p *StoreCachedMetadataProvider) get(ctx context.Context, current int) (mds *metadata.Metadata, data []byte, err error) {
	if data, err = p.latest(ctx, current); err != nil {
		return nil, nil, fmt.Errorf("error loading latest metadata: %w", err)
	}

	if data == nil {
		return nil, nil, nil
	}

	if mds, err = p.parse(bytes.NewReader(data)); err != nil {
		return nil, nil, fmt.Errorf("error parsing metadata: %w", err)
	}

	return mds, data, nil
}

func (p *StoreCachedMetadataProvider) LoadFile(ctx context.Context, path string) (mds *metadata.Metadata, data []byte, err error) {
	if data, err = os.ReadFile(path); err != nil {
		return nil, nil, fmt.Errorf("error reading file '%s': %w", path, err)
	}

	p.mu.Lock()

	defer p.mu.Unlock()

	if mds, err = p.parse(bytes.NewReader(data)); err != nil {
		return nil, nil, fmt.Errorf("error parsing metadata: %w", err)
	}

	if err = p.configure(mds); err != nil {
		return nil, nil, err
	}

	return mds, data, nil
}

func (p *StoreCachedMetadataProvider) LoadCache(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.loadCache(ctx)
}

func (p *StoreCachedMetadataProvider) loadCache(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	if mds, data, err = p.getCache(ctx); err != nil {
		return nil, nil, err
	}

	if mds == nil {
		return nil, nil, nil
	}

	if err = p.configure(mds); err != nil {
		return nil, nil, err
	}

	return mds, data, nil
}

func (p *StoreCachedMetadataProvider) getCache(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
	var cache *model.CachedData

	if cache, err = p.store.LoadCachedData(ctx, cacheMDS3); err != nil {
		return nil, nil, fmt.Errorf("error loading metadata cache from database: %w", err)
	}

	if cache == nil || cache.Value == nil {
		return nil, nil, nil
	}

	if mds, err = p.parse(bytes.NewReader(cache.Value)); err != nil {
		return nil, nil, fmt.Errorf("error parsing metadata cache from database: %w", err)
	}

	return mds, cache.Value, nil
}

func (p *StoreCachedMetadataProvider) SaveCache(ctx context.Context, data []byte) (err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.saveCache(ctx, data)
}

func (p *StoreCachedMetadataProvider) saveCache(ctx context.Context, data []byte) (err error) {
	if len(data) == 0 {
		return fmt.Errorf("error saving metadata cache to database: data is empty")
	}

	cache := model.CachedData{Name: cacheMDS3, Value: data}

	if err = p.store.SaveCachedData(ctx, cache); err != nil {
		return fmt.Errorf("error saving metadata cache to database: %w", err)
	}

	return nil
}

func (p *StoreCachedMetadataProvider) configure(mds *metadata.Metadata) (err error) {
	if mds == nil {
		return nil
	}

	var provider metadata.Provider

	if provider, err = p.new(mds); err != nil {
		return err
	}

	p.Provider = provider

	p.update, p.number = mds.Parsed.NextUpdate, mds.Parsed.Number

	return nil
}

func (p *StoreCachedMetadataProvider) Outdated() bool {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.outdated()
}

func (p *StoreCachedMetadataProvider) outdated() bool {
	return p.clock.Now().After(p.update)
}

func (p *StoreCachedMetadataProvider) latest(ctx context.Context, current int) (data []byte, err error) {
	if p.handler == nil {
		p.handler = &productionMDS3Provider{}
	}

	return p.handler.FetchMDS3(ctx, current)
}

func (p *StoreCachedMetadataProvider) parse(reader io.Reader) (mds *metadata.Metadata, err error) {
	var payload *metadata.PayloadJSON

	if payload, err = p.decoder.Decode(reader); err != nil {
		return nil, err
	}

	if mds, err = p.decoder.Parse(payload); err != nil {
		return nil, err
	}

	return mds, nil
}

type MDS3Provider interface {
	FetchMDS3(ctx context.Context, current int) (data []byte, err error)
}

type productionMDS3Provider struct {
	client *http.Client
}

func (h *productionMDS3Provider) FetchMDS3(ctx context.Context, current int) (data []byte, err error) {
	if h.client == nil {
		h.client = &http.Client{}
	}

	var (
		req  *http.Request
		resp *http.Response
	)

	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, metadata.ProductionMDSURL, nil); err != nil {
		return nil, fmt.Errorf("error creating request while attempting to get latest metadata from metadata service: %w", err)
	}

	if current > 0 {
		req.Header.Set(fasthttp.HeaderIfNoneMatch, fmt.Sprintf("%d", current))
	}

	if resp, err = h.client.Do(req); err != nil {
		return nil, fmt.Errorf("error getting latest metadata from metadata service: %w", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	case http.StatusNotModified:
		return nil, nil
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("error getting latest metadata from metadata service: too many requests")
	default:
		return nil, fmt.Errorf("error getting latest metadata from metadata service: unexpected status code: %d", resp.StatusCode)
	}
}
