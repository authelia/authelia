package webauthn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
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
		new:   newMetadataProviderMemory(config),
		clock: &metadata.RealClock{},
		store: store,
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
}

type StoreCachedMetadataProvider struct {
	metadata.Provider

	new func(mds *metadata.Metadata) (provider metadata.Provider, err error)

	mu      sync.RWMutex
	store   storage.CachedDataProvider
	decoder *metadata.Decoder
	clock   metadata.Clock
	client  *http.Client

	latestNextUpdate time.Time
	latestNumber     int
}

func (p *StoreCachedMetadataProvider) StartupCheck() (err error) {
	return p.init()
}

func (p *StoreCachedMetadataProvider) GetEntry(ctx context.Context, aaguid uuid.UUID) (entry *metadata.Entry, err error) {
	if err = p.update(ctx); err != nil {
		return nil, err
	}

	return p.Provider.GetEntry(ctx, aaguid)
}

func (p *StoreCachedMetadataProvider) init() (err error) {
	if p.store == nil {
		return fmt.Errorf("error initializing provider: storage is nil")
	}

	if err = p.update(context.Background()); err != nil {
		return err
	}

	return nil
}

func (p *StoreCachedMetadataProvider) update(ctx context.Context) (err error) {
	var (
		mds  *metadata.Metadata
		data []byte
	)

	p.mu.Lock()

	defer p.mu.Unlock()

	if mds, err = p.cached(); err == nil && mds != nil && !p.outdated() {
		if p.equal(mds) {
			return nil
		}
	} else {
		if data, err = p.latest(ctx); err != nil {
			return fmt.Errorf("error loading latest metdaata: %w", err)
		}

		if mds, err = p.parse(bytes.NewReader(data)); err != nil {
			return fmt.Errorf("error parsing metdaata: %w", err)
		}

		cache := model.CachedData{Name: cacheMDS3, Value: data}

		if err = p.store.SaveCachedData(ctx, cache); err != nil {
			return fmt.Errorf("error saving metadata cache to database: %w", err)
		}
	}

	var provider metadata.Provider

	if provider, err = p.new(mds); err != nil {
		return fmt.Errorf("error initializing metadata provider: %w", err)
	}

	p.Provider = provider

	p.latestNextUpdate, p.latestNumber = mds.Parsed.NextUpdate, mds.Parsed.Number

	return nil
}

func (p *StoreCachedMetadataProvider) equal(mds *metadata.Metadata) bool {
	return mds.Parsed.NextUpdate.Unix() == p.latestNextUpdate.Unix() && mds.Parsed.Number == p.latestNumber
}

func (p *StoreCachedMetadataProvider) outdated() bool {
	return p.clock.Now().After(p.latestNextUpdate)
}

func (p *StoreCachedMetadataProvider) cached() (mds *metadata.Metadata, err error) {
	var cache *model.CachedData

	if cache, err = p.store.LoadCachedData(context.Background(), cacheMDS3); err != nil {
		return nil, fmt.Errorf("error loading metadata cache from database: %w", err)
	}

	if cache == nil || cache.Value == nil {
		return nil, nil
	}

	if mds, err = p.parse(bytes.NewReader(cache.Value)); err != nil {
		return nil, fmt.Errorf("error parsing metadata cache from database: %w", err)
	}

	return mds, nil
}

func (p *StoreCachedMetadataProvider) latest(ctx context.Context) (data []byte, err error) {
	if p.client == nil {
		p.client = &http.Client{}
	}

	var (
		req  *http.Request
		resp *http.Response
	)

	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, metadata.ProductionMDSURL, nil); err != nil {
		return nil, fmt.Errorf("error creating request while attempting to get latest metadata from metadata service: %w", err)
	}

	if resp, err = p.client.Do(req); err != nil {
		return nil, fmt.Errorf("error getting latest metadata from metadata service: %w", err)
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
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
