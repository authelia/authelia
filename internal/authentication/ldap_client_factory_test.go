package authentication

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestPooledLDAPClientFactoryShouldNotDialPerRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 2},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	for range config.Pooling.Count {
		client := NewMockLDAPClient(ctrl)

		client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
		client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
		client.EXPECT().Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).Return(nil).Times(1)
		client.EXPECT().IsClosing().Return(false).AnyTimes()
		client.EXPECT().Close().Return(nil).AnyTimes()

		dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(client, nil).Times(1)
	}

	factory := NewPooledLDAPClientFactory(config, nil, dialer)

	require.NoError(t, factory.Initialize())

	for range 10 {
		acquired, err := factory.GetClient(WithPermitUnauthenticatedBind(config.PermitUnauthenticatedBind))

		require.NoError(t, err)
		require.IsType(t, &PooledLDAPClient{}, acquired)
		require.NoError(t, factory.ReleaseClient(acquired))
	}

	require.NoError(t, factory.Close())
}

func TestPooledLDAPClientFactoryGetClient(t *testing.T) {
	testCases := []struct {
		Name   string
		Opts   []LDAPClientFactoryOption
		Pooled bool
	}{
		{
			Name:   "ShouldPoolWithoutOptions",
			Opts:   nil,
			Pooled: true,
		},
		{
			Name:   "ShouldPoolWithPermitUnauthenticatedBind",
			Opts:   []LDAPClientFactoryOption{WithPermitUnauthenticatedBind(false)},
			Pooled: true,
		},
		{
			Name:   "ShouldPoolWithPermitUnauthenticatedBindEnabled",
			Opts:   []LDAPClientFactoryOption{WithPermitUnauthenticatedBind(true)},
			Pooled: true,
		},
		{
			Name:   "ShouldPoolWithMatchingAddress",
			Opts:   []LDAPClientFactoryOption{WithAddress(testLDAPURL)},
			Pooled: true,
		},
		{
			Name:   "ShouldNotPoolWithReferralAddress",
			Opts:   []LDAPClientFactoryOption{WithAddress("ldap://127.0.0.1:390"), WithPermitUnauthenticatedBind(false)},
			Pooled: false,
		},
		{
			Name:   "ShouldNotPoolWithUserCredentials",
			Opts:   []LDAPClientFactoryOption{WithUsername("uid=john,ou=users,dc=example,dc=com"), WithPassword("userpassword")},
			Pooled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config := &schema.AuthenticationBackendLDAP{
				Address:  testLDAPAddress,
				User:     "cn=admin,dc=example,dc=com",
				Password: "password",
				BaseDN:   "dc=example,dc=com",
				Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 1},
			}

			dialer := NewMockLDAPClientDialer(ctrl)
			client := NewMockLDAPClient(ctrl)

			dialer.EXPECT().DialURL(gomock.Eq(testLDAPURL), gomock.Any()).Return(client, nil).Times(1)
			client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
			client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
			client.EXPECT().Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).Return(nil).Times(1)
			client.EXPECT().IsClosing().Return(false).AnyTimes()
			client.EXPECT().Close().Return(nil).AnyTimes()

			factory := NewPooledLDAPClientFactory(config, nil, dialer)

			require.NoError(t, factory.Initialize())

			if !tc.Pooled {
				unpooled := NewMockLDAPClient(ctrl)

				dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(unpooled, nil).Times(1)
				unpooled.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
				unpooled.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
				unpooled.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				unpooled.EXPECT().Close().Return(nil).Times(1)
			}

			acquired, err := factory.GetClient(tc.Opts...)

			require.NoError(t, err)
			require.NotNil(t, acquired)

			_, pooled := acquired.(*PooledLDAPClient)

			assert.Equal(t, tc.Pooled, pooled)

			require.NoError(t, factory.ReleaseClient(acquired))
			require.NoError(t, factory.Close())
		})
	}
}

func TestPooledLDAPClientFactoryClose(t *testing.T) {
	testCases := []struct {
		Name       string
		Initialize bool
		Setup      func(t *testing.T, factory *PooledLDAPClientFactory)
	}{
		{
			Name:       "ShouldNotPanicWhenCalledTwice",
			Initialize: true,
			Setup: func(t *testing.T, factory *PooledLDAPClientFactory) {
				require.NoError(t, factory.Close())
			},
		},
		{
			Name:       "ShouldNotPanicWhenNotInitialized",
			Initialize: false,
			Setup:      nil,
		},
		{
			Name:       "ShouldNotPanicAcquiringAfterClose",
			Initialize: true,
			Setup: func(t *testing.T, factory *PooledLDAPClientFactory) {
				require.NoError(t, factory.Close())

				client, err := factory.GetClient()

				assert.Nil(t, client)
				assert.EqualError(t, err, "error acquiring client: the pool is closed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config := &schema.AuthenticationBackendLDAP{
				Address:  testLDAPAddress,
				User:     "cn=admin,dc=example,dc=com",
				Password: "password",
				BaseDN:   "dc=example,dc=com",
				Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 1},
			}

			dialer := NewMockLDAPClientDialer(ctrl)

			if tc.Initialize {
				client := NewMockLDAPClient(ctrl)

				client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
				client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
				client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				client.EXPECT().IsClosing().Return(false).AnyTimes()
				client.EXPECT().Close().Return(nil).AnyTimes()

				dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).AnyTimes()
			}

			factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

			require.True(t, ok)

			if tc.Initialize {
				require.NoError(t, factory.Initialize())
			}

			if tc.Setup != nil {
				tc.Setup(t, factory)
			}

			require.NotPanics(t, func() {
				require.NoError(t, factory.Close())
			})
		})
	}
}

func TestPooledLDAPClientFactoryShouldBackfillUnderFilledPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 2},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	gomock.InOrder(
		dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).Times(1),
		dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(nil, errors.New("server unavailable")).Times(1),
	)

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)
	require.NoError(t, factory.Initialize())

	assert.Equal(t, 1, factory.size)

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).Times(1)

	first, err := factory.GetClient()

	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := factory.GetClient()

	require.NoError(t, err)
	require.NotNil(t, second)

	assert.Equal(t, 2, factory.size)

	require.NoError(t, factory.ReleaseClient(first))
	require.NoError(t, factory.ReleaseClient(second))
	require.NoError(t, factory.Close())
}

func TestPooledLDAPClientFactoryShouldRefreshDiscoveryOnHealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dialed := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "",
				Attributes: []*ldap.EntryAttribute{{Name: ldapSupportedLDAPVersionAttribute, Values: []string{"2"}}},
			},
		},
	}

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 1},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	gomock.InOrder(
		client.EXPECT().Search(gomock.Any()).Return(dialed, nil).Times(1),
		client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes(),
	)

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).Times(1)

	factory := NewPooledLDAPClientFactory(config, nil, dialer)

	require.NoError(t, factory.Initialize())

	acquired, err := factory.GetClient()

	require.NoError(t, err)
	require.NotNil(t, acquired)

	assert.Equal(t, []int{3}, acquired.Discovery().LDAPVersion)

	require.NoError(t, factory.ReleaseClient(acquired))
	require.NoError(t, factory.Close())
}

func TestPooledLDAPClientFactoryShouldFailFastOnCloseWhileAcquiring(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling: schema.AuthenticationBackendLDAPPooling{
			Count:   1,
			Retries: 1,
			Timeout: time.Minute,
		},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).AnyTimes()

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)
	require.NoError(t, factory.Initialize())

	held, err := factory.GetClient()

	require.NoError(t, err)
	require.NotNil(t, held)

	acquired, started := make(chan error), make(chan struct{})

	go func() {
		close(started)

		_, aerr := factory.GetClient()

		acquired <- aerr
	}()

	<-started

	// The goroutine must reach the blocking select inside acquire before the pool is closed, otherwise it returns via
	// the closing check at the top of acquire and the blocking path goes untested.
	time.Sleep(time.Millisecond * 100)

	// Close blocks until the held client is released, so it runs concurrently with the assertion that the waiting
	// acquire is woken.
	closed := make(chan error, 1)

	go func() {
		closed <- factory.Close()
	}()

	select {
	case err = <-acquired:
		assert.EqualError(t, err, "error acquiring client: the pool is closed")
	case <-time.After(time.Second * 10):
		t.Fatal("acquire did not return promptly after the pool was closed")
	}

	require.NoError(t, factory.ReleaseClient(held))

	select {
	case err = <-closed:
		require.NoError(t, err)
	case <-time.After(time.Second * 10):
		t.Fatal("close did not return promptly after the held client was released")
	}
}

func TestPooledLDAPClientFactoryShouldNotMutateConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling:  schema.AuthenticationBackendLDAPPooling{},
	}

	factory, ok := NewPooledLDAPClientFactory(config, nil, NewMockLDAPClientDialer(ctrl)).(*PooledLDAPClientFactory)

	require.True(t, ok)

	assert.Equal(t, 0, config.Pooling.Count)
	assert.Equal(t, 0, config.Pooling.Retries)
	assert.Equal(t, time.Duration(0), config.Pooling.Timeout)

	expected := schema.DefaultLDAPAuthenticationBackendConfigurationImplementationCustom.Pooling

	assert.Equal(t, expected.Count, factory.count)
	assert.Equal(t, expected.Timeout, factory.timeout)
	assert.Equal(t, expected.Timeout/time.Duration(expected.Retries), factory.sleep)
}

func TestPooledLDAPClientFactoryInitialize(t *testing.T) {
	testCases := []struct {
		Name  string
		Dials int
		Setup func(t *testing.T, factory *PooledLDAPClientFactory)
		Err   string
	}{
		{
			Name:  "ShouldBeIdempotent",
			Dials: 2,
			Setup: func(t *testing.T, factory *PooledLDAPClientFactory) {
				require.NoError(t, factory.Initialize())
			},
			Err: "",
		},
		{
			Name:  "ShouldErrorAfterClose",
			Dials: 0,
			Setup: func(t *testing.T, factory *PooledLDAPClientFactory) {
				require.NoError(t, factory.Close())
			},
			Err: "error initializing client pool: the pool is closed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			config := &schema.AuthenticationBackendLDAP{
				Address:  testLDAPAddress,
				User:     "cn=admin,dc=example,dc=com",
				Password: "password",
				BaseDN:   "dc=example,dc=com",
				Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 2},
			}

			dialer := NewMockLDAPClientDialer(ctrl)

			client := NewMockLDAPClient(ctrl)

			client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
			client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
			client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			client.EXPECT().IsClosing().Return(false).AnyTimes()
			client.EXPECT().Close().Return(nil).AnyTimes()

			dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).Times(tc.Dials)

			factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

			require.True(t, ok)

			if tc.Setup != nil {
				tc.Setup(t, factory)
			}

			err := factory.Initialize()

			if tc.Err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.Dials, factory.size)
			} else {
				assert.EqualError(t, err, tc.Err)
			}
		})
	}
}

func TestPooledLDAPClientFactoryShouldInitializeConcurrentlyWithoutRacing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling:  schema.AuthenticationBackendLDAPPooling{Count: 4},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).Times(config.Pooling.Count)

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)

	var wg sync.WaitGroup

	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			assert.NoError(t, factory.Initialize())
		}()
	}

	wg.Wait()

	assert.Equal(t, config.Pooling.Count, factory.size)
	assert.Len(t, factory.pool, config.Pooling.Count)

	require.NoError(t, factory.Close())
}

func TestPooledLDAPClientFactoryCloseShouldWaitForCheckedOutClients(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling: schema.AuthenticationBackendLDAPPooling{
			Count:   2,
			Retries: 1,
			Timeout: time.Minute,
		},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).AnyTimes()

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)
	require.NoError(t, factory.Initialize())

	first, err := factory.GetClient()

	require.NoError(t, err)

	second, err := factory.GetClient()

	require.NoError(t, err)

	closed := make(chan error, 1)

	go func() {
		closed <- factory.Close()
	}()

	time.Sleep(time.Millisecond * 100)

	select {
	case <-closed:
		t.Fatal("close returned before the checked out clients were released")
	default:
	}

	require.NoError(t, factory.ReleaseClient(first))

	time.Sleep(time.Millisecond * 100)

	select {
	case <-closed:
		t.Fatal("close returned before all checked out clients were released")
	default:
	}

	require.NoError(t, factory.ReleaseClient(second))

	select {
	case err = <-closed:
		require.NoError(t, err)
	case <-time.After(time.Second * 10):
		t.Fatal("close did not return after all checked out clients were released")
	}

	assert.Equal(t, 0, factory.size)
}

func TestPooledLDAPClientFactoryCloseShouldTimeoutOnUnreleasedClients(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling: schema.AuthenticationBackendLDAPPooling{
			Count:   1,
			Retries: 1,
			Timeout: time.Millisecond * 250,
		},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).Return(client, nil).AnyTimes()

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)
	require.NoError(t, factory.Initialize())

	leaked, err := factory.GetClient()

	require.NoError(t, err)
	require.NotNil(t, leaked)

	err = factory.Close()

	assert.EqualError(t, err, "errors occurred closing the client pool: timeout of 250ms elapsed waiting for 1 checked out clients to be released")
}

func TestPooledLDAPClientFactoryCloseShouldNotLeakClientsDialedByInitialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling: schema.AuthenticationBackendLDAPPooling{
			Count:   1,
			Retries: 1,
			Timeout: time.Second * 5,
		},
	}

	var (
		mu     sync.Mutex
		closes int
	)

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().DoAndReturn(func() error {
		mu.Lock()

		closes++

		mu.Unlock()

		return nil
	}).AnyTimes()

	dialing, proceed := make(chan struct{}), make(chan struct{})

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).DoAndReturn(func(addr string, opts ...ldap.DialOpt) (LDAPBaseClient, error) {
		close(dialing)

		<-proceed

		return client, nil
	}).Times(1)

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)

	initialized := make(chan error, 1)

	go func() {
		initialized <- factory.Initialize()
	}()

	<-dialing

	require.NoError(t, factory.Close())

	close(proceed)

	select {
	case err := <-initialized:
		assert.EqualError(t, err, "error initializing client pool: the pool is closed")
	case <-time.After(time.Second * 10):
		t.Fatal("initialize did not return")
	}

	mu.Lock()

	assert.Equal(t, 1, closes)

	mu.Unlock()

	assert.Equal(t, 0, factory.size)
	assert.Equal(t, 0, len(factory.pool))
}

func TestPooledLDAPClientFactoryCloseShouldNotStallWhenGrowingFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &schema.AuthenticationBackendLDAP{
		Address:  testLDAPAddress,
		User:     "cn=admin,dc=example,dc=com",
		Password: "password",
		BaseDN:   "dc=example,dc=com",
		Pooling: schema.AuthenticationBackendLDAPPooling{
			Count:   2,
			Retries: 1,
			Timeout: time.Second * 5,
		},
	}

	dialer := NewMockLDAPClientDialer(ctrl)

	client := NewMockLDAPClient(ctrl)

	client.EXPECT().SetTimeout(gomock.Any()).AnyTimes()
	client.EXPECT().Search(gomock.Any()).Return(testLDAPRootDSESearchResult(), nil).AnyTimes()
	client.EXPECT().Bind(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	client.EXPECT().IsClosing().Return(false).AnyTimes()
	client.EXPECT().Close().Return(nil).AnyTimes()

	var (
		mu    sync.Mutex
		dials int
	)

	dialing, proceed := make(chan struct{}), make(chan struct{})

	dialer.EXPECT().DialURL(gomock.Any(), gomock.Any()).DoAndReturn(func(addr string, opts ...ldap.DialOpt) (LDAPBaseClient, error) {
		mu.Lock()

		dials++

		n := dials

		mu.Unlock()

		switch n {
		case 1:
			return client, nil
		case 2:
			return nil, errors.New("dial failed")
		default:
			close(dialing)

			<-proceed

			return nil, errors.New("dial failed")
		}
	}).AnyTimes()

	factory, ok := NewPooledLDAPClientFactory(config, nil, dialer).(*PooledLDAPClientFactory)

	require.True(t, ok)
	require.NoError(t, factory.Initialize())

	held, err := factory.GetClient()

	require.NoError(t, err)

	grown := make(chan error, 1)

	go func() {
		_, errGrow := factory.GetClient()

		grown <- errGrow
	}()

	<-dialing

	closed := make(chan error, 1)

	go func() {
		closed <- factory.Close()
	}()

	time.Sleep(time.Millisecond * 100)

	require.NoError(t, factory.ReleaseClient(held))

	time.Sleep(time.Millisecond * 100)

	select {
	case <-closed:
		t.Fatal("close returned before the growing client was accounted for")
	default:
	}

	close(proceed)

	select {
	case err = <-closed:
		require.NoError(t, err)
	case <-time.After(time.Second * 2):
		t.Fatal("close stalled after the pool failed to grow")
	}

	assert.Error(t, <-grown)
	assert.Equal(t, 0, factory.size)
}

func testLDAPRootDSESearchResult() *ldap.SearchResult {
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "",
				Attributes: []*ldap.EntryAttribute{
					{Name: ldapSupportedLDAPVersionAttribute, Values: []string{"3"}},
				},
			},
		},
	}
}
