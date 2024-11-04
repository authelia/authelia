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

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

// NewMetaDataProvider generates a new metadata.Provider given a *schema.Configuration and storage.CachedDataProvider.
func NewMetaDataProvider(config *schema.Configuration, store storage.CachedDataProvider) (provider MetaDataProvider, err error) {
	p := &StoreCachedMetadataProvider{
		new:     newMetadataProviderMemory(config),
		clock:   &metadata.RealClock{},
		store:   store,
		handler: &productionMDS3Provider{},
	}

	if p.decoder, err = metadata.NewDecoder(metadata.WithIgnoreEntryParsingErrors()); err != nil {
		return nil, err
	}

	return p, nil
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
	Load(ctx context.Context) (err error)
	LoadFile(ctx context.Context, path string) (err error)
	LoadCache(ctx context.Context) (err error)
	SaveCache(ctx context.Context) (err error)
	Outdated() (outdated bool)
}

type StoreCachedMetadataProvider struct {
	metadata.Provider

	data []byte

	new func(mds *metadata.Metadata) (provider metadata.Provider, err error)

	mu      sync.Mutex
	store   storage.CachedDataProvider
	decoder *metadata.Decoder
	clock   metadata.Clock
	handler MDS3Provider

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
		if err = p.loadCache(ctx); err != nil {
			if err = p.load(ctx); err != nil {
				return nil, err
			}
		}
	}

	return p.Provider.GetEntry(ctx, aaguid)
}

func (p *StoreCachedMetadataProvider) init() (err error) {
	if p.store == nil {
		return fmt.Errorf("error initializing provider: storage is not configured")
	}

	ctx := context.Background()

	if err = p.loadCache(ctx); err == nil && !p.outdated() {
		return nil
	}

	if err = p.load(ctx); err != nil {
		return err
	}

	return nil
}

func (p *StoreCachedMetadataProvider) LoadFile(ctx context.Context, path string) (err error) {
	var (
		data []byte
		mds  *metadata.Metadata
	)

	if data, err = os.ReadFile(path); err != nil {
		return fmt.Errorf("error reading file '%s': %w", path, err)
	}

	p.mu.Lock()

	defer p.mu.Unlock()

	if mds, err = p.parse(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("error parsing metdaata: %w", err)
	}

	return p.configure(mds, data)
}

func (p *StoreCachedMetadataProvider) LoadCache(ctx context.Context) (err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.loadCache(ctx)
}

func (p *StoreCachedMetadataProvider) loadCache(ctx context.Context) (err error) {
	var (
		mds  *metadata.Metadata
		data []byte
	)

	if mds, data, err = p.cached(ctx); err != nil {
		return err
	}

	if mds == nil {
		return nil
	}

	return p.configure(mds, data)
}

func (p *StoreCachedMetadataProvider) Load(ctx context.Context) (err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.load(ctx)
}

func (p *StoreCachedMetadataProvider) load(ctx context.Context) (err error) {
	var (
		mds  *metadata.Metadata
		data []byte
	)

	if data, err = p.latest(ctx); err != nil {
		return fmt.Errorf("error loading latest metdaata: %w", err)
	}

	if mds, err = p.parse(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("error parsing metdaata: %w", err)
	}

	return p.configure(mds, data)
}

func (p *StoreCachedMetadataProvider) SaveCache(ctx context.Context) (err error) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.save(ctx)
}

func (p *StoreCachedMetadataProvider) save(ctx context.Context) (err error) {
	cache := model.CachedData{Name: cacheMDS3, Value: p.data}

	if err = p.store.SaveCachedData(ctx, cache); err != nil {
		return fmt.Errorf("error saving metadata cache to database: %w", err)
	}

	return nil
}

func (p *StoreCachedMetadataProvider) configure(mds *metadata.Metadata, data []byte) (err error) {
	var provider metadata.Provider

	if provider, err = p.new(mds); err != nil {
		return fmt.Errorf("error initializing metadata provider: %w", err)
	}

	p.Provider = provider
	p.data = data

	p.update, p.number = mds.Parsed.NextUpdate, mds.Parsed.Number

	return nil
}

func (p *StoreCachedMetadataProvider) equal(mds *metadata.Metadata) bool {
	return mds.Parsed.NextUpdate.Unix() == p.update.Unix() && mds.Parsed.Number == p.number
}

func (p *StoreCachedMetadataProvider) Outdated() bool {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.outdated()
}

func (p *StoreCachedMetadataProvider) outdated() bool {
	return p.clock.Now().After(p.update)
}

func (p *StoreCachedMetadataProvider) cached(ctx context.Context) (mds *metadata.Metadata, data []byte, err error) {
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

func (p *StoreCachedMetadataProvider) latest(ctx context.Context) (data []byte, err error) {
	if p.handler == nil {
		p.handler = &productionMDS3Provider{}
	}

	return p.handler.FetchMDS3(ctx)
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
	FetchMDS3(ctx context.Context) (data []byte, err error)
}

type productionMDS3Provider struct {
	client *http.Client
}

func (h *productionMDS3Provider) FetchMDS3(ctx context.Context) (data []byte, err error) {
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

	if resp, err = h.client.Do(req); err != nil {
		return nil, fmt.Errorf("error getting latest metadata from metadata service: %w", err)
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
